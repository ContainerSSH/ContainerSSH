package sshserver_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

    "go.containerssh.io/libcontainerssh/auth"
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/internal/structutils"
    "go.containerssh.io/libcontainerssh/internal/test"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/metadata"
    "go.containerssh.io/libcontainerssh/message"
    "go.containerssh.io/libcontainerssh/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

//region Tests

func TestReadyRejection(t *testing.T) {
	//t.Parallel()()
	cfg := config.SSHConfig{}
	structutils.Defaults(&cfg)
	if err := cfg.GenerateHostKey(); err != nil {
		assert.Fail(t, "failed to generate host key", err)
		return
	}
	logger := log.NewTestLogger(t)
	handler := &rejectHandler{}

	server, err := sshserver.New(cfg, handler, logger)
	if err != nil {
		assert.Fail(t, "failed to create server", err)
		return
	}
	lifecycle := service.NewLifecycle(server)
	err = lifecycle.Run()
	if err == nil {
		assert.Fail(t, "server.Run() did not result in an error")
	} else {
		assert.Equal(t, "rejected", err.Error())
	}
	lifecycle.Stop(context.Background())
}

func TestAuthFailed(t *testing.T) {
	//t.Parallel()()
	port := test.GetNextPort(t, "SSH")
	server := newServerHelper(
		t,
		fmt.Sprintf("127.0.0.1:%d", port),
		map[string][]byte{
			"foo": []byte("bar"),
		},
		map[string]string{},
	)
	hostKey, err := server.start(t)
	if err != nil {
		assert.Fail(t, "failed to start ssh server", err)
		return
	}
	defer func() {
		server.stop()
		<-server.shutdownChannel
	}()

	sshConfig := &ssh.ClientConfig{
		User: "foo",
		Auth: []ssh.AuthMethod{ssh.Password("invalid")},
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		marshaledKey := key.Marshal()
		if bytes.Equal(marshaledKey, hostKey) {
			return nil
		}
		return fmt.Errorf("invalid host")
	}

	sshConnection, err := ssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), sshConfig)
	if err != nil {
		if !strings.Contains(err.Error(), "unable to authenticate") {
			assert.Fail(t, "handshake failed for non-auth reasons", err)
		}
	} else {
		_ = sshConnection.Close()
		assert.Fail(t, "authentication succeeded", err)
	}
}

func TestAuthKeyboardInteractive(t *testing.T) {
	//t.Parallel()()
	user1 := sshserver.NewTestUser("test")
	user1.AddKeyboardInteractiveChallengeResponse("foo", "bar")

	user2 := sshserver.NewTestUser("test")
	user2.AddKeyboardInteractiveChallengeResponse("foo", "baz")

	logger := log.NewTestLogger(t)
	srv := sshserver.NewTestServer(
		t,
		sshserver.NewTestAuthenticationHandler(
			sshserver.NewTestHandler(),
			user2,
		),
		logger,
		nil,
	)
	srv.Start()

	client1 := sshserver.NewTestClient(srv.GetListen(), srv.GetHostKey(), user1, logger)
	conn, err := client1.Connect()
	if err == nil {
		_ = conn.Close()
		t.Fatal("invalid keyboard-interactive authentication did not result in an error")
	}

	client2 := sshserver.NewTestClient(srv.GetListen(), srv.GetHostKey(), user2, logger)
	conn, err = client2.Connect()
	if err != nil {
		t.Fatalf("valid keyboard-interactive authentication resulted in an error (%v)", err)
	}
	_ = conn.Close()

	defer srv.Stop(10 * time.Second)
}

type gssApiClient struct {
	username string
}

func (c *gssApiClient) InitSecContext(target string, token []byte, isGSSDelegCreds bool) (outputToken []byte, needContinue bool, err error) {
	if token == nil {
		return []byte(c.username), true, nil
	} else {
		if string(token) != c.username {
			return []byte{}, false, fmt.Errorf("Invalid test token, expecting username")
		}
		return []byte{}, false, nil
	}
}

func (c *gssApiClient) GetMIC(micField []byte) ([]byte, error) {
	return append(micField, []byte(c.username)...), nil
}

func (c *gssApiClient) DeleteSecContext() error {
	return nil
}

// Test the GSSAPI plumbing within the sshserver
func TestAuthGSSAPI(t *testing.T) {
	user2 := sshserver.NewTestUser("foo")
	logger := log.NewTestLogger(t)

	sshconf := config.SSHConfig{}
	structutils.Defaults(&sshconf)

	srv := sshserver.NewTestServer(
		t,
		sshserver.NewTestAuthenticationHandler(
			sshserver.NewTestHandler(),
			user2,
		),
		logger,
		nil,
	)
	srv.Start()

	gssClient := gssApiClient{
		username: "foo",
	}
	sshConfig := &ssh.ClientConfig{
		User: "foo",
		Auth: []ssh.AuthMethod{ssh.GSSAPIWithMICAuthMethod(&gssClient, "testing.containerssh.io")},
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		marshaledKey := key.Marshal()
		private, err := ssh.ParsePrivateKey([]byte(srv.GetHostKey()))
		if err != nil {
			panic(err)
		}
		if bytes.Equal(marshaledKey, private.PublicKey().Marshal()) {
			return nil
		}
		return fmt.Errorf("invalid host")
	}

	sshConnection, err := ssh.Dial("tcp", srv.GetListen(), sshConfig)
	if err != nil {
		if !strings.Contains(err.Error(), "unable to authenticate") {
			assert.Fail(t, "handshake failed for non-auth reasons", err)
		}
	} else {
		_ = sshConnection.Close()
		assert.Fail(t, "authentication succeeded", err)
	}

	defer srv.Stop(10 * time.Second)
}

func TestSessionSuccess(t *testing.T) {
	//t.Parallel()()
	port := test.GetNextPort(t, "SSH")
	server := newServerHelper(
		t,
		fmt.Sprintf("127.0.0.1:%d", port),
		map[string][]byte{
			"foo": []byte("bar"),
		},
		map[string]string{},
	)
	hostKey, err := server.start(t)
	if err != nil {
		assert.Fail(t, "failed to start ssh server", err)
		return
	}
	defer func() {
		server.stop()
		<-server.shutdownChannel
	}()

	reply, exitStatus, err := shellRequestReply(
		fmt.Sprintf("127.0.0.1:%d", port),
		"foo",
		ssh.Password("bar"),
		hostKey,
		[]byte("Hi"),
		nil,
		nil,
	)
	assert.Equal(t, []byte("Hello world!"), reply)
	assert.Equal(t, 0, exitStatus)
	assert.Equal(t, nil, err)
}

func TestSessionError(t *testing.T) {
	//t.Parallel()()
	port := test.GetNextPort(t, "SSH")
	server := newServerHelper(
		t,
		fmt.Sprintf("127.0.0.1:%d", port),
		map[string][]byte{
			"foo": []byte("bar"),
		},
		map[string]string{},
	)
	hostKey, err := server.start(t)
	if err != nil {
		assert.Fail(t, "failed to start ssh server", err)
		return
	}
	defer func() {
		server.stop()
		<-server.shutdownChannel
	}()

	reply, exitStatus, err := shellRequestReply(
		fmt.Sprintf("127.0.0.1:%d", port),
		"foo",
		ssh.Password("bar"),
		hostKey,
		[]byte("Ho"),
		nil,
		nil,
	)
	assert.Equal(t, 1, exitStatus)
	assert.Equal(t, []byte{}, reply)
	assert.Equal(t, nil, err)
}

func TestPubKey(t *testing.T) {
	//t.Parallel()()
	port := test.GetNextPort(t, "SSH")
	rsaKey, err := rsa.GenerateKey(
		rand.Reader,
		4096,
	)
	assert.Nil(t, err, "failed to generate RSA key (%v)", err)
	signer, err := ssh.NewSignerFromKey(rsaKey)
	assert.Nil(t, err, "failed to create signer (%v)", err)
	publicKey := signer.PublicKey()
	authorizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(publicKey)))
	server := newServerHelper(
		t,
		fmt.Sprintf("127.0.0.1:%d", port),
		map[string][]byte{},
		map[string]string{
			"foo": authorizedKey,
		},
	)
	hostKey, err := server.start(t)
	if err != nil {
		assert.Fail(t, "failed to start ssh server", err)
		return
	}
	defer func() {
		server.stop()
		<-server.shutdownChannel
	}()

	reply, exitStatus, err := shellRequestReply(
		fmt.Sprintf("127.0.0.1:%d", port),
		"foo",
		ssh.PublicKeys(signer),
		hostKey,
		[]byte("Hi"),
		nil,
		nil,
	)
	assert.Nil(t, err, "failed to send shell request (%v)", err)
	assert.Equal(t, 0, exitStatus)
	assert.Equal(t, []byte("Hello world!"), reply)
}

func TestKeepAlive(t *testing.T) {
	//t.Parallel()()

	logger := log.NewTestLogger(t)

	user := sshserver.NewTestUser("test")
	user.RandomPassword()

	config := config.SSHConfig{}
	structutils.Defaults(&config)
	config.ClientAliveInterval = 1 * time.Second
	srv := sshserver.NewTestServer(
		t,
		sshserver.NewTestAuthenticationHandler(
			sshserver.NewTestHandler(),
			user,
		),
		logger,
		&config,
	)
	srv.Start()
	defer srv.Stop(1 * time.Minute)

	hostkey, err := ssh.ParsePrivateKey([]byte(srv.GetHostKey()))
	if err != nil {
		t.Fatal("Failed to parse private key")
	}
	sshConfig := &ssh.ClientConfig{
		User: user.Username(),
		Auth: user.GetAuthMethods(),
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if bytes.Equal(key.Marshal(), hostkey.PublicKey().Marshal()) {
			return nil
		}
		return fmt.Errorf("invalid host")
	}
	tcpConnection, err := net.Dial("tcp", srv.GetListen())
	if err != nil {
		t.Fatal("tcp handshake failed (%w)", err)
	}
	connection, _, globalReq, err := ssh.NewClientConn(tcpConnection, srv.GetListen(), sshConfig)
	if err != nil {
		t.Fatal("ssh handshake failed (%w)", err)
	}
	defer func() {
		_ = connection.Close()
	}()

	req := <-globalReq
	err = req.Reply(false, nil)
	if err != nil {
		t.Fatal("Failed to respond to first request")
	}
	recv1 := time.Now()

	req2 := <-globalReq
	recv2 := time.Now()
	err = req.Reply(false, nil)
	if err != nil {
		t.Fatal("Failed to respond to second request")
	}

	if req.Type != "keepalive@openssh.com" {
		t.Fatal("Expected keepalive request", req.Type)
	}
	if req2.Type != "keepalive@openssh.com" {
		t.Fatal("Expected keepalive request", req.Type)
	}

	elapsed := recv2.Sub(recv1)

	if elapsed > 2*time.Second {
		t.Fatal("Received keepalive in too big of an interval", elapsed)
	}
	if elapsed < time.Second/2 {
		t.Fatal("Received keepalive in too short of an interval", elapsed)
	}
}

//endregion

//region Helper

func shellRequestReply(
	host string,
	user string,
	authMethod ssh.AuthMethod,
	hostKey []byte,
	request []byte,
	onShell chan struct{},
	canSendResponse chan struct{},
) (reply []byte, exitStatus int, err error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{authMethod},
	}
	sshConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if bytes.Equal(key.Marshal(), hostKey) {
			return nil
		}
		return fmt.Errorf("invalid host")
	}
	sshConnection, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, -1, fmt.Errorf("handshake failed (%w)", err)
	}
	defer func() {
		if sshConnection != nil {
			_ = sshConnection.Close()
		}
	}()

	session, err := sshConnection.NewSession()
	if err != nil {
		return nil, -1, fmt.Errorf("new session failed (%w)", err)
	}

	stdin, stdout, err := createPipe(session)
	if err != nil {
		return nil, -1, err
	}

	if err := session.Setenv("TERM", "xterm"); err != nil {
		return nil, -1, err
	}

	if err := session.Shell(); err != nil {
		return nil, -1, fmt.Errorf("failed to request shell (%w)", err)
	}
	if onShell != nil {
		onShell <- struct{}{}
	}
	if canSendResponse != nil {
		<-canSendResponse
	}
	if _, err := stdin.Write(request); err != nil {
		return nil, -1, fmt.Errorf("failed to write to shell (%w)", err)
	}
	return read(stdout, stdin, session)
}

func read(stdout io.Reader, stdin io.WriteCloser, session *ssh.Session) (
	[]byte,
	int,
	error,
) {
	var exitStatus int
	data := make([]byte, 4096)
	n, err := stdout.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, -1, fmt.Errorf("failed to read from stdout (%w)", err)
	}
	if err := stdin.Close(); err != nil && !errors.Is(err, io.EOF) {
		return data[:n], -1, fmt.Errorf("failed to close stdin (%w)", err)
	}
	if err := session.Wait(); err != nil {
		exitError := &ssh.ExitError{}
		if errors.As(err, &exitError) {
			exitStatus = exitError.ExitStatus()
		} else {
			return data[:n], -1, fmt.Errorf("failed to wait for exit (%w)", err)
		}
	}
	if err := session.Close(); err != nil && !errors.Is(err, io.EOF) {
		return data[:n], -1, fmt.Errorf("failed to close session (%w)", err)
	}
	return data[:n], exitStatus, nil
}

func createPipe(session *ssh.Session) (io.WriteCloser, io.Reader, error) {
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to request stdin (%w)", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to request stdout (%w)", err)
	}
	return stdin, stdout, nil
}

func newServerHelper(
	t *testing.T,
	listen string,
	passwords map[string][]byte,
	pubKeys map[string]string,
) *serverHelper {
	return &serverHelper{
		t:               t,
		listen:          listen,
		passwords:       passwords,
		pubKeys:         pubKeys,
		receivedChannel: make(chan struct{}, 1),
	}
}

type serverHelper struct {
	t               *testing.T
	server          sshserver.Server
	lifecycle       service.Lifecycle
	passwords       map[string][]byte
	pubKeys         map[string]string
	listen          string
	shutdownChannel chan struct{}
	receivedChannel chan struct{}
}

func (h *serverHelper) start(t *testing.T) (hostKey []byte, err error) {
	if h.server != nil {
		return nil, fmt.Errorf("server already running")
	}
	cfg := config.SSHConfig{}
	structutils.Defaults(&cfg)
	cfg.Listen = h.listen
	if err := cfg.GenerateHostKey(); err != nil {
		return nil, err
	}
	private, err := ssh.ParsePrivateKey([]byte(cfg.HostKeys[0]))
	if err != nil {
		return nil, err
	}
	hostKey = private.PublicKey().Marshal()
	logger := log.NewTestLogger(t)
	readyChannel := make(chan struct{}, 1)
	h.shutdownChannel = make(chan struct{}, 1)
	errChannel := make(chan error, 1)
	handler := newFullHandler(
		readyChannel,
		h.shutdownChannel,
		h.passwords,
		h.pubKeys,
	)
	server, err := sshserver.New(cfg, handler, logger)
	if err != nil {
		return hostKey, err
	}
	lifecycle := service.NewLifecycle(server)
	h.lifecycle = lifecycle
	go func() {
		err = lifecycle.Run()
		if err != nil {
			errChannel <- err
		}
	}()
	//Wait for the server to be ready
	select {
	case err := <-errChannel:
		return hostKey, err
	case <-readyChannel:
	}
	h.server = server
	return hostKey, nil
}

func (h *serverHelper) stop() {
	if h.lifecycle != nil {
		shutdownContext, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
		h.lifecycle.Stop(shutdownContext)
		cancelFunc()
	}
}

//endregion

//region Handlers

//region Rejection

type rejectHandler struct {
}

func (r *rejectHandler) OnReady() error {
	return fmt.Errorf("rejected")
}

func (r *rejectHandler) OnShutdown(_ context.Context) {
}

func (r *rejectHandler) OnNetworkConnection(meta metadata.ConnectionMetadata) (
	sshserver.NetworkConnectionHandler,
	metadata.ConnectionMetadata,
	error,
) {
	return nil, meta, fmt.Errorf("not implemented")
}

//endregion

//region Full

func newFullHandler(
	readyChannel chan struct{},
	shutdownChannel chan struct{},
	passwords map[string][]byte,
	pubKeys map[string]string,
) sshserver.Handler {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &fullHandler{
		ctx:          ctx,
		cancelFunc:   cancelFunc,
		ready:        readyChannel,
		shutdownDone: shutdownChannel,
		passwords:    passwords,
		pubKeys:      pubKeys,
	}
}

//region Handler

type fullHandler struct {
	sshserver.AbstractHandler

	ctx             context.Context
	shutdownContext context.Context
	cancelFunc      context.CancelFunc
	passwords       map[string][]byte
	pubKeys         map[string]string
	ready           chan struct{}
	shutdownDone    chan struct{}
}

func (f *fullHandler) OnReady() error {
	f.ready <- struct{}{}
	return nil
}

func (f *fullHandler) OnShutdown(shutdownContext context.Context) {
	f.shutdownContext = shutdownContext
	<-f.shutdownContext.Done()
	close(f.shutdownDone)
}

func (f *fullHandler) OnNetworkConnection(meta metadata.ConnectionMetadata) (
	sshserver.NetworkConnectionHandler,
	metadata.ConnectionMetadata,
	error,
) {
	return &fullNetworkConnectionHandler{
		handler: f,
	}, meta, nil
}

//endregion

//region Network connection conformanceTestHandler

type fullNetworkConnectionHandler struct {
	sshserver.AbstractNetworkConnectionHandler

	handler *fullHandler
}

func (f *fullNetworkConnectionHandler) OnAuthPassword(
	meta metadata.ConnectionAuthPendingMetadata,
	password []byte,
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	if storedPassword, ok := f.handler.passwords[meta.Username]; ok && bytes.Equal(storedPassword, password) {
		return sshserver.AuthResponseSuccess, meta.Authenticated(meta.Username), nil
	}
	return sshserver.AuthResponseFailure, meta.AuthFailed(), fmt.Errorf("authentication failed")
}

func (f *fullNetworkConnectionHandler) OnAuthPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	pubKey auth.PublicKey,
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	if storedPubKey, ok := f.handler.pubKeys[meta.Username]; ok && storedPubKey == pubKey.PublicKey {
		return sshserver.AuthResponseSuccess, meta.Authenticated(meta.Username), nil
	}
	return sshserver.AuthResponseFailure, meta.AuthFailed(), fmt.Errorf("authentication failed")
}

func (f *fullNetworkConnectionHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	sshserver.SSHConnectionHandler,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return &fullSSHConnectionHandler{
		handler: f.handler,
	}, meta, nil
}

//endregion

//region SSH connection conformanceTestHandler

type fullSSHConnectionHandler struct {
	sshserver.AbstractSSHConnectionHandler

	handler *fullHandler
}

func (f *fullSSHConnectionHandler) OnSessionChannel(
	_ metadata.ChannelMetadata,
	_ []byte,
	session sshserver.SessionChannel,
) (channel sshserver.SessionChannelHandler, failureReason sshserver.ChannelRejection) {
	return &fullSessionChannelHandler{
		handler: f.handler,
		env:     map[string]string{},
		session: session,
	}, nil
}

func (s *fullSSHConnectionHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwading channel unimplemented", "Forwading channel unimplemented")
}

func (s *fullSSHConnectionHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *fullSSHConnectionHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *fullSSHConnectionHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwading channel unimplemented in docker backend", "Forwading channel unimplemented in docker backend")
}

func (s *fullSSHConnectionHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *fullSSHConnectionHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	return fmt.Errorf("Unimplemented")
}

//endregion

//region Session channel conformanceTestHandler

type fullSessionChannelHandler struct {
	sshserver.AbstractSessionChannelHandler

	handler *fullHandler
	env     map[string]string
	session sshserver.SessionChannel
}

func (f *fullSessionChannelHandler) OnEnvRequest(_ uint64, name string, value string) error {
	f.env[name] = value
	return nil
}

func (f *fullSessionChannelHandler) OnShell(
	_ uint64,
) error {
	stdin := f.session.Stdin()
	stdout := f.session.Stdout()
	go func() {
		data := make([]byte, 4096)
		n, err := stdin.Read(data)
		if err != nil {
			f.session.ExitStatus(1)
			_ = f.session.Close()
			return
		}
		if string(data[:n]) != "Hi" {
			f.session.ExitStatus(1)
			_ = f.session.Close()
			return
		}
		if _, err := stdout.Write([]byte("Hello world!")); err != nil {
			f.session.ExitStatus(1)
			_ = f.session.Close()
		}
		f.session.ExitStatus(0)
		_ = f.session.Close()
	}()
	return nil
}

func (f *fullSessionChannelHandler) OnSignal(_ uint64, _ string) error {
	return nil
}

func (f *fullSessionChannelHandler) OnWindow(_ uint64, _ uint32, _ uint32, _ uint32, _ uint32) error {
	return nil
}

func (s *fullSessionChannelHandler) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return nil
}

func (f *fullSessionChannelHandler) OnShutdown(_ context.Context) {
	_ = f.session.Close()
}

//endregion

//endregion

//endregion
