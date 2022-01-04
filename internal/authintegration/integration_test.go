package authintegration_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/authintegration"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestAuthentication(t *testing.T) {
	logger := log.NewTestLogger(t)

	authServerPort := test.GetNextPort(t, "auth server")

	authLifecycle := startAuthServer(t, logger, authServerPort)
	defer authLifecycle.Stop(context.Background())

	sshServerConfig, lifecycle := startSSHServer(t, logger, authServerPort)
	defer lifecycle.Stop(context.Background())

	testConnection(t, ssh.Password("bar"), sshServerConfig, true)
	testConnection(t, ssh.Password("baz"), sshServerConfig, false)
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
	handler, _, err := authintegration.New(
		config.AuthConfig{
			Method: config.AuthMethodWebhook,
			Webhook: config.AuthWebhookClientConfig{
				HTTPClientConfiguration: config.HTTPClientConfiguration{
					URL:     fmt.Sprintf("http://127.0.0.1:%d", authServerPort),
					Timeout: 10 * time.Second,
				},
				Password: true,
				PubKey:   false,
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

func testConnection(t *testing.T, authMethod ssh.AuthMethod, sshServerConfig config.SSHConfig, success bool) {
	clientConfig := ssh.ClientConfig{
		Config: ssh.Config{},
		User:   "foo",
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

func (t *testBackend) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {}

func (t *testBackend) OnSessionChannel(_ uint64, _ []byte, _ sshserver.SessionChannel) (
	_ sshserver.SessionChannelHandler,
	_ sshserver.ChannelRejection,
) {
	return nil, sshserver.NewChannelRejection(ssh.UnknownChannelType, message.MTest, "not supported", "not supported")
}

func (t *testBackend) OnHandshakeFailed(_ error) {}

func (t *testBackend) OnHandshakeSuccess(_ string, _ string, _ *auth2.ConnectionMetadata) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	return t, nil
}

func (t *testBackend) OnDisconnect() {

}

func (t *testBackend) OnReady() error {
	return nil
}

func (t *testBackend) OnShutdown(_ context.Context) {

}

func (t *testBackend) OnNetworkConnection(_ net.TCPAddr, _ string) (
	sshserver.NetworkConnectionHandler,
	error,
) {
	return t, nil
}

// endregion

// region AuthHandler
type authHandler struct {
}

func (h *authHandler) OnPassword(
	Username string,
	Password []byte,
	_ string,
	_ string,
) (bool, *auth2.ConnectionMetadata, error) {
	if Username == "foo" && string(Password) == "bar" {
		return true, nil, nil
	}
	if Username == "crash" {
		// Simulate a database failure
		return false, nil, fmt.Errorf("database error")
	}
	return false, nil, nil
}

func (h *authHandler) OnPubKey(
	_ string,
	_ string,
	_ string,
	_ string,
) (bool, *auth2.ConnectionMetadata, error) {
	return false, nil, nil
}

func (h *authHandler) OnAuthorization(
	_ string,
	_ string,
	_ string,
	_ string,
) (bool, *auth2.ConnectionMetadata, error) {
	return false, nil, nil
}

// endregion
