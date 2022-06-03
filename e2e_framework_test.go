package libcontainerssh_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

    containerssh "go.containerssh.io/libcontainerssh"
    auth2 "go.containerssh.io/libcontainerssh/auth"
    "go.containerssh.io/libcontainerssh/auth/webhook"
    "go.containerssh.io/libcontainerssh/config"
    internalssh "go.containerssh.io/libcontainerssh/internal/ssh"
    "go.containerssh.io/libcontainerssh/internal/test"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/metadata"
    "go.containerssh.io/libcontainerssh/service"
	"golang.org/x/crypto/ssh"
)

func NewT(t *testing.T) T {
	cfg := config.AppConfig{}
	cfg.Default()
	return &testContext{
		t,
		cfg,
		&sync.Mutex{},
		NewAuthUserStorage(),
		0,
		0,
		nil,
		nil,
		nil,
		nil,
	}
}

type T interface {
	// StartContainerSSH starts a ContainerSSH instance on a random port. If no authentication has been previously
	// configured, this configures ContainerSSH for webhook authentication with an internal user database.
	StartContainerSSH()
	// ConfigureBackend configures ContainerSSH to use a specific backend.
	ConfigureBackend(backend config.Backend)
	// LoginViaSSH logs creates a temporary user and logs in via SSH. After this
	// has been called new session channels can be requested.
	LoginViaSSH()
	// StartSessionChannel starts a new session channel in ContainerSSH. Later commands will
	// run in the context of this session channel. If multiple session channels are desired,
	// each one should be run in a separate subtest.
	StartSessionChannel()
	// RequestCommandExecution attempts to run the specified command in a previously-opened
	// session channel. If no session channel has been opened, this command will fail.
	RequestCommandExecution(cmd string)
	// RequestShell requests a shell to be executed.
	RequestShell()
	// AssertStdoutHas waits for the specified output string to be sent from the output.
	AssertStdoutHas(output string)
	// SendStdin sends the specified string to the SSH server via the standard input.
	SendStdin(data string)
	// CloseChannel closes the current channel.
	CloseChannel()

	Parallel()
	Run(name string, f func(t T)) bool
	Cleanup(func())
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
}

type testContext struct {
	*testing.T
	cfg       config.AppConfig
	lock      *sync.Mutex
	users     AuthUserStorage
	authPort  int
	sshPort   int
	lifecycle service.Lifecycle
	sshConn   *ssh.Client
	channel   ssh.Channel
	requests  <-chan *ssh.Request
}

func (c *testContext) StartContainerSSH() {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.lifecycle != nil {
		t.Fatalf("ContainerSSH is already running.")
	}

	t.Logf("Starting ContainerSSH...")

	c.authPort = test.GetNextPort(c.T, "auth server")
	c.authServer(c.T, c.users, c.authPort)
	c.cfg.Auth.PasswordAuth.Method = config.PasswordAuthMethodWebhook
	c.cfg.Auth.PasswordAuth.Webhook.URL = fmt.Sprintf("http://127.0.0.1:%d", c.authPort)
	c.cfg.Auth.PublicKeyAuth.Method = config.PubKeyAuthMethodWebhook
	c.cfg.Auth.PublicKeyAuth.Webhook.URL = fmt.Sprintf("http://127.0.0.1:%d", c.authPort)

	c.sshPort = test.GetNextPort(c.T, "ContainerSSH")
	c.cfg.SSH.Listen = fmt.Sprintf("127.0.0.1:%d", c.sshPort)
	if err := c.cfg.SSH.GenerateHostKey(); err != nil {
		t.Fatalf("Failed to generate host keys (%v)", err)
	}
	c.cfg.Log.T = c.T
	c.cfg.Log.Destination = config.LogDestinationTest

	cssh, lifecycle, err := containerssh.New(c.cfg, log.NewLoggerFactory())
	if err != nil {
		t.Fatalf("Failed to start ContainerSSH (%v)", err)
	}
	c.lifecycle = lifecycle
	running := make(chan struct{})
	stopped := make(chan struct{})
	crashed := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			close(running)
		},
	)
	lifecycle.OnStopping(
		func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
			close(stopped)
		},
	)
	lifecycle.OnCrashed(
		func(s service.Service, l service.Lifecycle, err error) {
			close(crashed)
		},
	)
	go func() {
		_ = cssh.RunWithLifecycle(lifecycle)
	}()
	c.T.Cleanup(func() {
		lifecycle.Stop(context.Background())
	})

	select {
	case <-running:
		t.Logf("ContainerSSH is now running.")
	case <-stopped:
		t.Fatalf("ContainerSSH unexpectedly stopped.")
	case <-crashed:
		t.Fatalf("ContainerSSH unexpectedly crashed.")
	}

	t.Logf("Started ContainerSSH.")
}

func (c *testContext) ConfigureBackend(backend config.Backend) {
	t := c.T
	t.Helper()

	t.Logf("Configuring %s backend...", backend)
	c.cfg.Backend = backend
	switch backend {
	case config.BackendKubernetes:
		kube := test.Kubernetes(t)
		c.cfg.Kubernetes.Connection.ServerName = kube.ServerName
		c.cfg.Kubernetes.Connection.Host = kube.Host
		c.cfg.Kubernetes.Connection.CAData = kube.CACert
		c.cfg.Kubernetes.Connection.KeyData = kube.UserKey
		c.cfg.Kubernetes.Connection.CertData = kube.UserCert
	case config.BackendSSHProxy:
		proxy := test.SSH(t)
		c.cfg.SSHProxy.Server = proxy.Host()
		c.cfg.SSHProxy.Port = uint16(proxy.Port())
		c.cfg.SSHProxy.Username = proxy.Username()
		c.cfg.SSHProxy.Password = proxy.Password()
		c.cfg.SSHProxy.AllowedHostKeyFingerprints = []string{
			proxy.FingerprintSHA256(),
		}
		c.cfg.SSHProxy.HostKeyAlgorithms = config.MustSSHKeyAlgoListFromStringList(
			proxy.HostKeyAlgorithms(),
		)
	}
	t.Logf("Configured %s backend.", backend)
}

func (c *testContext) LoginViaSSH() {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sshConn != nil {
		t.Fatalf("Already logged in via SSH.")
	}

	t.Logf("Logging in via SSH...")

	cfg := ssh.ClientConfig{}
	cfg.SetDefaults()
	username := c.T.Name()
	user := c.users.AddUser(username)
	password := "test-login"
	user.SetPassword(password)
	cfg.Auth = append(cfg.Auth, ssh.Password(password))

	hostKeys, err := c.cfg.SSH.LoadHostKeys()
	if err != nil {
		t.Fatalf("Failed to read back the host keys (%v)", err)
	}
	hostKeyAlgorithms := make([]string, len(hostKeys))
	marshalledHostKeys := make([]string, len(hostKeys))
	for i, hostKey := range hostKeys {
		hostKeyAlgorithms[i] = hostKey.PublicKey().Type()
		marshalledHostKeys[i] = string(ssh.MarshalAuthorizedKey(hostKey.PublicKey()))
	}
	cfg.HostKeyAlgorithms = hostKeyAlgorithms
	cfg.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		marshalledHostKey := string(ssh.MarshalAuthorizedKey(key))
		for _, hostKey := range marshalledHostKeys {
			if hostKey == marshalledHostKey {
				return nil
			}
		}
		return fmt.Errorf("invalid host key: %s", marshalledHostKey)
	}
	cfg.User = username

	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", c.sshPort), &cfg)
	if err != nil {
		t.Fatalf("Failed to log in via SSH (%v)", err)
	}
	c.sshConn = sshConn

	t.Logf("SSH login successful.")
}

func (c *testContext) StartSessionChannel() {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sshConn == nil {
		t.Fatalf("SSH connection is not running.")
	}
	if c.channel != nil {
		t.Fatalf("A channel is already open.")
	}

	t.Logf("Starting a new session channel...")

	channel, requests, err := c.sshConn.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("Failed to open channel (%v)", err)
	}
	c.channel = channel
	c.requests = requests
	// We use c.T here so the cleanup happens in the parent test.
	c.T.Cleanup(func() {
		_ = channel.Close()
		c.channel = nil
	})

	t.Logf("Started a new session channel.")
}

func (c *testContext) RequestCommandExecution(cmd string) {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.channel == nil {
		t.Fatalf("No channel opened.")
	}
	t.Logf("Starting program %s ...", cmd)

	success, err := c.channel.SendRequest(
		string(internalssh.RequestTypeExec),
		true,
		ssh.Marshal(internalssh.ExecRequestPayload{Exec: cmd}),
	)
	if err != nil {
		t.Fatalf("Failed to send exec request. (%v)", err)
	}
	if !success {
		t.Fatalf("Server rejected exec request. (%v)", err)
	}
	t.Logf("Started program.")
}

func (c *testContext) RequestShell() {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.channel == nil {
		t.Fatalf("No channel opened.")
	}
	t.Logf("Starting shell...")

	success, err := c.channel.SendRequest(
		string(internalssh.RequestTypeShell),
		true,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to send exec request. (%v)", err)
	}
	if !success {
		t.Fatalf("Server rejected exec request. (%v)", err)
	}
	t.Logf("Started shell.")
}

func (c *testContext) AssertStdoutHas(output string) {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.channel == nil {
		t.Fatalf("No channel opened.")
	}
	t.Logf("Waiting for output...")

	// Wait a second to allow the output to arrive
	time.Sleep(1 * time.Second)
	data := make([]byte, 16*1024)
	n, err := c.channel.Read(data)
	if err != nil {
		t.Fatalf("Failed to read channel stdout. (%v)", err)
	}

	if !strings.Contains(string(data[:n]), output) {
		t.Fatalf("Output does not contain '%s' (output was: %s)", output, string(data[:n]))
	}

	t.Logf("Output check complete.")
}

func (c *testContext) SendStdin(data string) {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.channel == nil {
		t.Fatalf("No channel opened.")
	}
	t.Logf("Sending stdin data...")

	if _, err := c.channel.Write([]byte(data)); err != nil {
		t.Fatalf("Failed to send stdin data (%v)", err)
	}

	t.Logf("Sent stdin data.")
}

func (c *testContext) CloseChannel() {
	t := c.T
	t.Helper()
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.channel == nil {
		t.Fatalf("No channel opened.")
	}
	t.Logf("Closing channel...")

	if err := c.channel.Close(); err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Failed to close channel. (%v)", err)
	}

	c.channel = nil

	t.Logf("Closed channel.")
}

func (c *testContext) Run(name string, f func(t T)) bool {
	c.T.Helper()
	return c.T.Run(name, func(t *testing.T) {
		t.Helper()
		f(&testContext{
			t,
			c.cfg,
			c.lock,
			c.users,
			c.authPort,
			c.sshPort,
			c.lifecycle,
			c.sshConn,
			nil,
			nil,
		})
	})
}

type authHandler struct {
	userdb AuthUserStorage
}

func (a *authHandler) OnPassword(meta metadata.ConnectionAuthPendingMetadata, Password []byte) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	user, err := a.userdb.GetUser(meta.Username)
	if err != nil {
		return false, meta.AuthFailed(), err
	}
	if pw := user.GetPassword(); pw != nil && *pw == string(Password) {
		return true, meta.Authenticated(meta.Username), nil
	}
	return false, meta.AuthFailed(), fmt.Errorf("incorrect password")
}

func (a *authHandler) OnPubKey(meta metadata.ConnectionAuthPendingMetadata, publicKey auth2.PublicKey) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	user, err := a.userdb.GetUser(meta.Username)
	if err != nil {
		return false, meta.AuthFailed(), err
	}
	for _, key := range user.GetAuthorizedKeys() {
		if key == publicKey.PublicKey {
			return true, meta.Authenticated(meta.Username), nil
		}
	}
	return false, meta.AuthFailed(), fmt.Errorf("authentication failed")
}

func (a *authHandler) OnAuthorization(meta metadata.ConnectionAuthenticatedMetadata) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return true, meta, nil
}

func (c *testContext) authServer(t *testing.T, userdb AuthUserStorage, port int) {
	srv, err := webhook.NewServer(
		config.HTTPServerConfiguration{
			Listen: fmt.Sprintf("127.0.0.1:%d", port),
		},
		&authHandler{
			userdb: userdb,
		},
		log.NewTestLogger(t),
	)
	if err != nil {
		t.Fatalf("Failed to start authentication webhook server. (%v)", err)
	}
	lifecycle := service.NewLifecycle(srv)
	go func() {
		_ = lifecycle.Run()
	}()

	t.Cleanup(func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		lifecycle.Stop(shutdownContext)

		lastError := lifecycle.Wait()
		if lastError != nil {
			t.Fatalf("Failed to stop authentication webhook server. (%v)", lastError)
		}
	})
}

// AuthUser is an entry in the in-memory AuthUserStorage. It can be used to modify the user used for testing.
type AuthUser interface {
	SetPassword(password string)
	// GetPassword
	GetPassword() *string
	// AddKey adds a new private and public key to this user.
	AddKey() ssh.Signer
	// GetKeys returns a list of signers containing the private and public key for this user.
	GetKeys() []ssh.Signer
	// GetAuthorizedKeys returns a list of public keys in the OpenSSH Authorized Keys format.
	GetAuthorizedKeys() []string
}

// AuthUserStorage is a storage interface for creating and managing in-memory users used for test
// authentications.
type AuthUserStorage interface {
	// AddUser adds a new user to the in-memory database. You can then add credentials for the
	// SSH connection to the user. If the user already exists, this function throws a panic.
	AddUser(username string) AuthUser
	// GetUser returns a user with a specific username. If the user is not found, this function
	// throws a panic.
	GetUser(username string) (AuthUser, error)
	// RemoveUser removes a user from the internal database. If the user is not found, this
	// function throws a panic.
	RemoveUser(username string)
}

type authUserStorage struct {
	lock  *sync.Mutex
	users map[string]AuthUser
}

func (a *authUserStorage) AddUser(username string) AuthUser {
	a.lock.Lock()
	defer a.lock.Unlock()
	if _, ok := a.users[username]; ok {
		panic(message.NewMessage(message.MTest, "User %s already exists in test user database.", username))
	}
	a.users[username] = &authUser{
		lock: &sync.Mutex{},
	}
	return a.users[username]
}

func (a *authUserStorage) GetUser(username string) (AuthUser, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if user, ok := a.users[username]; ok {
		return user, nil
	}
	return nil, message.NewMessage(message.MTest, "User %s not found in test database.", username)
}

func (a *authUserStorage) RemoveUser(username string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.users, username)
}

type authUser struct {
	lock     *sync.Mutex
	password *string
	keys     []ssh.Signer
}

func (a *authUser) SetPassword(password string) {
	a.password = &password
}

func (a *authUser) GetPassword() *string {
	return a.password
}

func (a *authUser) AddKey() ssh.Signer {
	a.lock.Lock()
	defer a.lock.Unlock()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		panic(message.Wrap(err, message.MTest, "Failed to generate RSA key."))
	}
	var pemBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	var pemBytes bytes.Buffer
	if err := pem.Encode(&pemBytes, pemBlock); err != nil {
		panic(message.Wrap(err, message.MTest, "Failed to marshal private key."))
	}

	sshPrivateKey, err := ssh.ParsePrivateKey(pemBytes.Bytes())
	if err != nil {
		panic(message.Wrap(err, message.MTest, "Failed to parse SSH private key."))
	}

	a.keys = append(a.keys, sshPrivateKey)
	return sshPrivateKey
}

func (a *authUser) GetKeys() []ssh.Signer {
	a.lock.Lock()
	defer a.lock.Unlock()
	result := make([]ssh.Signer, len(a.keys))
	copy(result, a.keys)
	return result
}

func (a *authUser) GetAuthorizedKeys() []string {
	a.lock.Lock()
	defer a.lock.Unlock()
	result := make([]string, len(a.keys))
	for i, key := range a.keys {
		result[i] = fmt.Sprintf("ssh-rsa %s", ssh.MarshalAuthorizedKey(key.PublicKey()))
	}
	return result
}

// NewAuthUserStorage creates a new in-memory user storage for authentication.
func NewAuthUserStorage() AuthUserStorage {
	return &authUserStorage{
		lock:  &sync.Mutex{},
		users: map[string]AuthUser{},
	}
}
