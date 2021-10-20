package authintegration_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	error2 "github.com/containerssh/containerssh/error"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/geoip"
	"github.com/containerssh/http"
	"github.com/containerssh/metrics"
	"github.com/containerssh/service"
	sshserver "github.com/containerssh/sshserver/v2"
	"github.com/containerssh/structutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"

	"github.com/containerssh/authintegration"
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
		http.ServerConfiguration{
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
	sshserver.Config,
	service.Lifecycle,
) {
	backend := &testBackend{}
	geoipLookup, _ := geoip.New(geoip.Config{
		Provider: "dummy",
	})
	collector := metrics.New(geoipLookup)
	handler, _, err := authintegration.New(
		auth.ClientConfig{
			Method: auth.MethodWebhook,
			Webhook: auth.WebhookClientConfig{
				ClientConfiguration: http.ClientConfiguration{
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

	sshServerConfig := sshserver.Config{}
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

func testConnection(t *testing.T, authMethod ssh.AuthMethod, sshServerConfig sshserver.Config, success bool) {
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
	return nil, sshserver.NewChannelRejection(ssh.UnknownChannelType, error2.MTest, "not supported", "not supported")
}

func (t *testBackend) OnHandshakeFailed(_ error) {}

func (t *testBackend) OnHandshakeSuccess(_ string) (
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
) (bool, error) {
	if Username == "foo" && string(Password) == "bar" {
		return true, nil
	}
	if Username == "crash" {
		// Simulate a database failure
		return false, fmt.Errorf("database error")
	}
	return false, nil
}

func (h *authHandler) OnPubKey(
	_ string,
	_ string,
	_ string,
	_ string,
) (bool, error) {
	return false, nil
}

// endregion
