package auth_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/service"
	"github.com/stretchr/testify/assert"
)

type handler struct {
}

func (h *handler) OnPassword(
	username string,
	password []byte,
	remoteAddress string,
	connectionID string,
) (bool, *auth2.ConnectionMetadata, error) {
	if remoteAddress != "127.0.0.1" {
		return false, nil, fmt.Errorf("invalid IP: %s", remoteAddress)
	}
	if connectionID != "0123456789ABCDEF" {
		return false, nil, fmt.Errorf("invalid connection ID: %s", connectionID)
	}
	if username == "foo" && string(password) == "bar" {
		return true, nil, nil
	}
	if username == "crash" {
		// Simulate a database failure
		return false, nil, fmt.Errorf("database error")
	}
	return false, nil, nil
}

func (h *handler) OnPubKey(
	username string,
	publicKey string,
	remoteAddress string,
	connectionID string,
) (bool, *auth2.ConnectionMetadata, error) {
	if remoteAddress != "127.0.0.1" {
		return false, nil, fmt.Errorf("invalid IP: %s", remoteAddress)
	}
	if connectionID != "0123456789ABCDEF" {
		return false, nil, fmt.Errorf("invalid connection ID: %s", connectionID)
	}
	if username == "foo" && publicKey == "ssh-rsa asdf" {
		return true, nil, nil
	}
	if username == "crash" {
		// Simulate a database failure
		return false, nil, fmt.Errorf("database error")
	}
	return false, nil, nil
}

func (h *handler) OnAuthorization(
	principalUsername string,
	loginUsername string,
	remoteAddress string,
	connectionID string,
) (bool, *auth2.ConnectionMetadata, error) {
	return false, nil, nil
}

func TestAuth(t *testing.T) {
	logger := log.NewTestLogger(t)
	logger.Info(
		message.NewMessage(
			"TEST",
			"FYI: errors during this test are expected as we test against error cases.",
		),
	)

	for name, subpath := range map[string]string{"empty": "", "auth": "/auth"} {
		t.Run(fmt.Sprintf("subpath_%s", name), func(t *testing.T) {
			client, lifecycle, metricsCollector, err := initializeAuth(t, logger, subpath)
			if err != nil {
				assert.Fail(t, "failed to initialize auth", err)
				return
			}
			defer lifecycle.Stop(context.Background())

			authenticationContext := client.Password("foo", []byte("bar"), "0123456789ABCDEF", net.ParseIP("127.0.0.1"))
			assert.Equal(t, nil, authenticationContext.Error())
			assert.Equal(t, true, authenticationContext.Success())
			assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthBackendRequests)[0].Value)
			assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthSuccess)[0].Value)

			authenticationContext = client.Password("foo", []byte("baz"), "0123456789ABCDEF", net.ParseIP("127.0.0.1"))
			assert.Equal(t, nil, authenticationContext.Error())
			assert.Equal(t, false, authenticationContext.Success())
			assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthFailure)[0].Value)

			authenticationContext = client.Password("crash", []byte("baz"), "0123456789ABCDEF", net.ParseIP("127.0.0.1"))
			assert.NotEqual(t, nil, authenticationContext.Error())
			assert.Equal(t, false, authenticationContext.Success())
			assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthBackendFailure)[0].Value)

			authenticationContext = client.PubKey("foo", "ssh-rsa asdf", "0123456789ABCDEF", net.ParseIP("127.0.0.1"))
			assert.Equal(t, nil, authenticationContext.Error())
			assert.Equal(t, true, authenticationContext.Success())

			authenticationContext = client.PubKey("foo", "ssh-rsa asdx", "0123456789ABCDEF", net.ParseIP("127.0.0.1"))
			assert.Equal(t, nil, authenticationContext.Error())
			assert.Equal(t, false, authenticationContext.Success())

			authenticationContext = client.PubKey("crash", "ssh-rsa asdx", "0123456789ABCDEF", net.ParseIP("127.0.0.1"))
			assert.NotEqual(t, nil, authenticationContext.Error())
			assert.Equal(t, false, authenticationContext.Success())
		})
	}
}

func initializeAuth(t *testing.T, logger log.Logger, subpath string) (
	auth.Client,
	service.Lifecycle,
	metrics.Collector,
	error,
) {
	ready := make(chan bool, 1)
	errors := make(chan error)
	port := test.GetNextPort(t, "auth server")

	server, err := auth.NewServer(
		config.HTTPServerConfiguration{
			Listen: fmt.Sprintf("127.0.0.1:%d", port),
		},
		&handler{},
		logger,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	metricsCollector := metrics.New(dummy.New())

	client, err := auth.NewHttpAuthClient(
		config.AuthConfig{
			Method: config.AuthMethodWebhook,
			Webhook: config.AuthWebhookClientConfig{
				HTTPClientConfiguration: config.HTTPClientConfiguration{
					URL:     fmt.Sprintf("http://127.0.0.1:%d%s", port, subpath),
					Timeout: 2 * time.Second,
				},
				Password: true,
				PubKey:   true,
			},
			AuthTimeout: 2 * time.Second,
		},
		logger,
		metricsCollector,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	lifecycle := service.NewLifecycle(server)
	lifecycle.OnRunning(
		func(_ service.Service, _ service.Lifecycle) {
			ready <- true
		},
	)

	go func() {
		if err := lifecycle.Run(); err != nil {
			errors <- err
		}
		close(errors)
	}()
	<-ready
	return client, lifecycle, metricsCollector, nil
}
