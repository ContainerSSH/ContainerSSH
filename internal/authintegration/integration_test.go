package authintegration_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/auth"
	"github.com/containerssh/containerssh/internal/authintegration"
	"github.com/containerssh/containerssh/internal/geoip/dummy"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
	"github.com/containerssh/containerssh/service"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestAuthentication(t *testing.T) {
	logger := log.NewTestLogger(t)

	authLifecycle := startAuthServer(t, logger)
	defer authLifecycle.Stop(context.Background())

	sshServerConfig, lifecycle := startSSHServer(t, logger)
	defer lifecycle.Stop(context.Background())

	testConnection(t, ssh.Password("bar"), sshServerConfig, true)
	testConnection(t, ssh.Password("baz"), sshServerConfig, false)
}

func startAuthServer(t *testing.T, logger log.Logger) service.Lifecycle {
	server, err := auth.NewServer(
		config.HTTPServerConfiguration{
			Listen: "127.0.0.1:8080",
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

func startSSHServer(t *testing.T, logger log.Logger) (
	config.SSHConfig,
	service.Lifecycle,
) {
	backend := &testBackend{}
	collector := metrics.New(dummy.New())
	handler, _, err := authintegration.New(
		config.AuthConfig{
			Method: config.AuthMethodWebhook,
			Webhook: config.AuthWebhookClientConfig{
				HTTPClientConfiguration: config.HTTPClientConfiguration{
					URL:     "http://127.0.0.1:8080",
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
		Config:          ssh.Config{},
		User:            "foo",
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

func (t *testBackend) OnHandshakeSuccess(_ string, _ string, _ map[string]string) (
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
) (bool, map[string]string, error) {
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
) (bool, map[string]string, error) {
	return false, nil, nil
}

// endregion
