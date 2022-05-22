package auth_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"

	configuration "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/metadata"
	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/ssh"

	"github.com/containerssh/gokrb5/v8/client"
	krb5cfg "github.com/containerssh/gokrb5/v8/config"
	"github.com/containerssh/gokrb5/v8/crypto"
	"github.com/containerssh/gokrb5/v8/gssapi"
	"github.com/containerssh/gokrb5/v8/iana/flags"
	"github.com/containerssh/gokrb5/v8/messages"
	"github.com/containerssh/gokrb5/v8/spnego"
	"github.com/containerssh/gokrb5/v8/types"
)

func tempFile(t *testing.T) *os.File {
	file, err := ioutil.TempFile(t.TempDir(), "krb5.keytab-*")
	if err != nil {
		panic(err)
	}
	return file
}

func setupKerberosClient(
	t *testing.T,
	authType auth.AuthenticationType,
	config configuration.AuthKerberosClientConfig,
) auth.KerberosClient {
	krb := test.Kerberos(t)
	kt := tempFile(t)
	n, err := kt.Write(krb.Keytab())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(kt.Name())
		if err != nil {
			t.Fatal("Failed to remove keytab file", err)
		}
	}()
	if n != len(krb.Keytab()) {
		t.Fatal("Failed to write keytab")
	}
	config.Keytab = kt.Name()
	err = kt.Close()
	if err != nil {
		t.Fatal("Failed to close keytab file", err)
	}

	c, err := auth.NewKerberosClient(
		authType,
		config,
		log.NewTestLogger(t),
		metrics.New(dummy.New()),
	)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestKerberosPasswordAuth(t *testing.T) {
	config := configuration.AuthKerberosClientConfig{
		EnforceUsername: true,
		ConfigPath:      "testdata/krb5.conf",
	}
	c := setupKerberosClient(t, auth.AuthenticationTypePassword, config)

	ctx := c.Password(
		metadata.ConnectionAuthPendingMetadata{
			ConnectionMetadata: metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
			Username: "foo",
		},
		[]byte("test"),
	)
	if !ctx.Success() {
		assert.Fail(t, "Failed to login as foo", ctx.Error())
	}

	ctx = c.Password(
		metadata.ConnectionAuthPendingMetadata{
			ConnectionMetadata: metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
			Username: "foo",
		},
		[]byte("wrongpass"),
	)
	if ctx.Success() {
		assert.Fail(t, "Logged in as foo with wrong password")
	}
}

func TestKerberosNoPassword(t *testing.T) {
	config := configuration.AuthKerberosClientConfig{
		EnforceUsername: true,
		ConfigPath:      "testdata/krb5.conf",
	}
	c := setupKerberosClient(t, auth.AuthenticationTypeGSSAPI, config)

	ctx := c.Password(
		metadata.ConnectionAuthPendingMetadata{
			ConnectionMetadata: metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
			Username: "foo",
		},
		[]byte("test"),
	)
	if ctx.Success() {
		assert.Fail(t, "Logged in as foo even though password is disallowed")
	}

	ctx = c.Password(
		metadata.ConnectionAuthPendingMetadata{
			ConnectionMetadata: metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
			Username: "foo",
		},
		[]byte("wrongpass"),
	)
	if ctx.Success() {
		assert.Fail(t, "Logged in as foo with wrong password")
	}
}

func TestKerberosGSSAPIAuth(t *testing.T) {
	config := configuration.AuthKerberosClientConfig{}
	structutils.Defaults(&config)
	config.ConfigPath = "testdata/krb5.conf"

	authCl := setupKerberosClient(t, auth.AuthenticationTypeGSSAPI, config)
	userCl := krbAuth(t, "foo", "test")

	gssClient := testGssApiClient{
		client: userCl,
	}

	name, err := doGssAuth(
		t,
		"foo",
		"host/testing.containerssh.io",
		false,
		authCl.GSSAPI(
			metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
		),
		gssClient,
	)
	if err != nil {
		assert.Fail(t, "GSSAPI Failure", err)
	}
	if name != "foo" {
		assert.Fail(t, "Logged in with the wrong name", name)
	}

	name, err = doGssAuth(
		t,
		"foo",
		"host/testing.containerssh.io",
		true,
		authCl.GSSAPI(
			metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
		),
		gssClient,
	)
	if err != nil {
		assert.Fail(t, "GSSAPI Failure", err)
	}
	if name != "foo" {
		assert.Fail(t, "Logged in with the wrong name", name)
	}

	_, err = doGssAuth(
		t,
		"bar",
		"host/testing.containerssh.io",
		false,
		authCl.GSSAPI(
			metadata.ConnectionMetadata{
				RemoteAddress: metadata.RemoteAddress(
					net.TCPAddr{
						IP:   net.IPv4(127, 0, 0, 1),
						Port: 22,
					},
				),
				ConnectionID: "asdf",
				Metadata:     map[string]metadata.Value{},
				Environment:  map[string]metadata.Value{},
				Files:        map[string]metadata.BinaryValue{},
			},
		),
		gssClient,
	)
	if err == nil {
		assert.Fail(t, "Invalid username was successfully authenticated")
	}
}

func krbAuth(t *testing.T, username string, password string) *client.Client {
	conf, err := krb5cfg.Load("testdata/krb5.conf")
	if err != nil {
		panic(err)
	}

	cl := client.NewWithPassword(
		username,
		conf.LibDefaults.DefaultRealm,
		password,
		conf,
		client.DisablePAFXFAST(true),
	)

	err = cl.Login()
	if err != nil {
		t.Fatal("Failed to authenticate to kerbers server as foo", err)
	}
	return cl
}

func doGssAuth(
	t *testing.T,
	username string,
	principal string,
	deleg bool,
	server auth.GSSAPIServer,
	client testGssApiClient,
) (string, error) {
	tok, cont, err := client.InitSecContext(principal, nil, deleg)
	if err != nil {
		return "", fmt.Errorf("Client failed to initialize context (%w)", err)
	}
	if !cont {
		return "", fmt.Errorf("Client did not ask for any token exchanges (%w)", err)
	}

	_, name, cont, err := server.AcceptSecContext(tok)
	if err != nil {
		return "", fmt.Errorf("Server failed to initialize context (%w)", err)
	}
	if cont {
		return "", fmt.Errorf("Server asked for a second token exchange but we don't support this yet")
	}

	micField := buildMic(username)
	mic, err := client.GetMIC(micField)
	if err != nil {
		return "", fmt.Errorf("Failed to get MIC from client (%w)", err)
	}

	err = server.VerifyMIC(micField, mic)
	if err != nil {
		return "", fmt.Errorf("Server failed to verify mic (%w)", err)
	}

	err = server.DeleteSecContext()
	if err != nil {
		return "", fmt.Errorf("Server failed to delete sec context (%w)", err)
	}

	err = client.DeleteSecContext()
	if err != nil {
		return "", fmt.Errorf("Client failed to delete sec context (%w)", err)
	}

	return name, nil
}

func buildMic(username string) []byte {
	mic := auth.GSSAPIMicField{
		SessionIdentifier: "abcdefg",
		Request:           50,
		UserName:          username,
		Service:           "ssh-connection",
		Method:            "gssapi-with-mic",
	}

	return ssh.Marshal(mic)
}

type testGssApiClient struct {
	client *client.Client
	key    types.EncryptionKey
}

func (c *testGssApiClient) generateContext(target string, deleg bool) (outputToken []byte, needContinue bool, err error) {
	tkt, key, err := c.client.GetServiceTicket(target)
	if err != nil {
		return nil, false, err
	}

	tok, err := spnego.NewKRB5TokenAPREQ(
		c.client,
		tkt,
		key,
		[]int{
			gssapi.ContextFlagMutual,
			gssapi.ContextFlagInteg,
		},
		[]int{
			flags.APOptionMutualRequired,
		},
	)
	if err != nil {
		return nil, false, err
	}

	err = tok.APReq.DecryptAuthenticator(key)
	if err != nil {
		return nil, false, err
	}

	etype, err := crypto.GetEtype(key.KeyType)
	if err != nil {
		return nil, false, err
	}

	err = tok.APReq.Authenticator.GenerateSeqNumberAndSubKey(key.KeyType, etype.GetKeyByteSize())
	if err != nil {
		return nil, false, err
	}

	c.key = tok.APReq.Authenticator.SubKey

	req, err := messages.NewAPReq(tkt, key, tok.APReq.Authenticator)
	if err != nil {
		return nil, false, err
	}

	tok.APReq = req

	mar, err := tok.Marshal()
	if err != nil {
		return nil, false, err
	}

	return mar, true, nil
}

func (c *testGssApiClient) InitSecContext(target string, token []byte, isGSSDelegCreds bool) (outputToken []byte, needContinue bool, err error) {
	if token == nil {
		// Initial setup
		return c.generateContext(target, isGSSDelegCreds)
	} else {
		// Got a response
		var rep spnego.KRB5Token
		err := rep.Unmarshal(token)
		if err != nil {
			return nil, false, err
		}

		if rep.IsKRBError() {
			return nil, false, fmt.Errorf("Received error response")
		}

		if !rep.IsAPRep() {
			return nil, false, fmt.Errorf("Did not receive expected response packet AP_REP")
		}
		return nil, false, nil
	}
}

func (c *testGssApiClient) GetMIC(micField []byte) ([]byte, error) {
	token, err := gssapi.NewInitiatorMICToken(micField, c.key)
	if err != nil {
		return nil, err
	}

	mar, err := token.Marshal()
	if err != nil {
		return nil, err
	}
	return mar, nil
}

func (c *testGssApiClient) DeleteSecContext() error {
	return nil
}
