package authintegration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	publicAuth "go.containerssh.io/containerssh/auth"
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/auth"
	"go.containerssh.io/containerssh/internal/authintegration"
	"go.containerssh.io/containerssh/internal/geoip/dummy"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/internal/sshserver"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/internal/test"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/message"
	"go.containerssh.io/containerssh/metadata"
	"go.containerssh.io/containerssh/service"
	"golang.org/x/crypto/ssh"
)

func TestAuthentication(t *testing.T) {
	logger := log.NewTestLogger(t)

	authServerPort := test.GetNextPort(t, "auth server")

	authLifecycle := startAuthServer(t, logger, authServerPort)
	defer authLifecycle.Stop(context.Background())

	sshServerConfig, lifecycle := startSSHServer(t, logger, authServerPort)
	defer lifecycle.Stop(context.Background())

	testConnection(t, "foo", ssh.Password("bar"), sshServerConfig, true)
	testConnection(t, "foo", ssh.Password("baz"), sshServerConfig, false)

	testConnection(t, "foonoauthz", ssh.Password("bar"), sshServerConfig, false)
	testConnection(t, "foonoauthz", ssh.Password("baz"), sshServerConfig, false)
}

func startAuthServer(t *testing.T, logger log.Logger, authServerPort int) service.Lifecycle {
	server, err := auth.NewServer(
		config.HTTPServerConfiguration{
			Listen: fmt.Sprintf("127.0.0.1:%d", authServerPort),
		},
		&authHandler{},
		logger,
	)
	assert.NoError(t, err)

	lifecycle := service.NewLifecycle(server)
	ready := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			ready <- struct{}{}
		},
	)

	go func() {
		assert.NoError(t, lifecycle.Run())
	}()
	<-ready
	return lifecycle
}

func startSSHServer(t *testing.T, logger log.Logger, authServerPort int) (config.SSHConfig, service.Lifecycle) {
	backend := &testBackend{}
	collector := metrics.New(dummy.New())
	webhookConfig := config.AuthWebhookClientConfig{
		HTTPClientConfiguration: config.HTTPClientConfiguration{
			URL:     fmt.Sprintf("http://127.0.0.1:%d", authServerPort),
			Timeout: 10 * time.Second,
		},
		AuthTimeout: 30 * time.Second,
	}
	handler, _, err := authintegration.New(
		config.AuthConfig{
			PasswordAuth: config.PasswordAuthConfig{
				Method:  config.PasswordAuthMethodWebhook,
				Webhook: webhookConfig,
			},
			Authz: config.AuthzConfig{
				Method:  config.AuthzMethodWebhook,
				Webhook: webhookConfig,
			},
		},
		backend,
		logger,
		collector,
		authintegration.BehaviorNoPassthrough,
	)
	assert.NoError(t, err)

	sshServerConfig := config.SSHConfig{}
	structutils.Defaults(&sshServerConfig)
	assert.NoError(t, sshServerConfig.GenerateHostKey())
	srv, err := sshserver.New(
		sshServerConfig,
		handler,
		logger,
	)
	assert.NoError(t, err)

	lifecycle := service.NewLifecycle(srv)

	ready := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			ready <- struct{}{}
		},
	)

	go func() {
		assert.NoError(t, lifecycle.Run())
	}()
	<-ready
	return sshServerConfig, lifecycle
}

func testConnection(t *testing.T, username string, authMethod ssh.AuthMethod, sshServerConfig config.SSHConfig, success bool) {
	clientConfig := ssh.ClientConfig{
		Config: ssh.Config{},
		User:   username,
		Auth:   []ssh.AuthMethod{authMethod},
		// We don't care about host key verification for this test.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}
	client, err := ssh.Dial("tcp", sshServerConfig.Listen, &clientConfig)
	if success {
		assert.Nil(t, err)
	} else {
		assert.NotNil(t, err)
	}
	if client != nil {
		assert.NoError(t, client.Close())
	}
}

// region Backend
type testBackend struct {
	sshserver.AbstractNetworkConnectionHandler
}

func (t *testBackend) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {}

func (b *testBackend) OnFailedDecodeGlobalRequest(_ uint64, _ string, _ []byte, _ error) {}

func (t *testBackend) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {}

func (t *testBackend) OnSessionChannel(_ metadata.ChannelMetadata, _ []byte, _ sshserver.SessionChannel) (
	_ sshserver.SessionChannelHandler,
	_ sshserver.ChannelRejection,
) {
	return nil, sshserver.NewChannelRejection(ssh.UnknownChannelType, message.MTest, "not supported", "not supported")
}

func (s *testBackend) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwarding channel unimplemented in docker backend", "Forwarding channel unimplemented in docker backend")
}

func (s *testBackend) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testBackend) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testBackend) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "Forwarding channel unimplemented in docker backend", "Forwarding channel unimplemented in docker backend")
}

func (s *testBackend) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (s *testBackend) OnRequestCancelStreamLocal(
	path string,
) error {
	return fmt.Errorf("Unimplemented")
}

func (t *testBackend) OnHandshakeFailed(_ metadata.ConnectionMetadata, _ error) {}

func (t *testBackend) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	return t, meta, nil
}

func (t *testBackend) OnDisconnect() {

}

func (t *testBackend) OnReady() error {
	return nil
}

func (t *testBackend) OnShutdown(_ context.Context) {

}

func (t *testBackend) OnNetworkConnection(meta metadata.ConnectionMetadata) (
	sshserver.NetworkConnectionHandler,
	metadata.ConnectionMetadata,
	error,
) {
	return t, meta, nil
}

func (t *testBackend) OnRequestAuthAgent(reverseHandler sshserver.ReverseForward) error {
	return fmt.Errorf("Unimplemented")
}

func (t *testBackend) OnAuthAgentChannel(channelID uint64) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	return nil, sshserver.NewChannelRejection(ssh.Prohibited, message.ESSHNotImplemented, "SSH agent channel unimplemented in test backend", "SSH agent channel unimplemented in test backend")
}

// endregion

// region AuthHandler
type authHandler struct {
}

func (h *authHandler) OnPassword(
	meta metadata.ConnectionAuthPendingMetadata,
	Password []byte,
) (bool, metadata.ConnectionAuthenticatedMetadata, error) {
	if (meta.Username == "foo" || meta.Username == "foonoauthz") && string(Password) == "bar" {
		return true, meta.Authenticated(meta.Username), nil
	}
	if meta.Username == "crash" {
		// Simulate a database failure
		return false, meta.AuthFailed(), fmt.Errorf("database error")
	}
	return false, meta.AuthFailed(), nil
}

func (h *authHandler) OnPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	_ publicAuth.PublicKey,
) (bool, metadata.ConnectionAuthenticatedMetadata, error) {
	return false, meta.AuthFailed(), nil
}

func (h *authHandler) OnAuthorization(
	meta metadata.ConnectionAuthenticatedMetadata,
) (bool, metadata.ConnectionAuthenticatedMetadata, error) {
	if meta.AuthenticatedUsername == "foo" {
		return true, meta, nil
	}
	return false, meta.AuthFailed(), nil
}

// endregion
