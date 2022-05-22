package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	auth3 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"
	"github.com/containerssh/libcontainerssh/service"
)

type handler struct {
}

func (h *handler) OnPassword(meta metadata.ConnectionAuthPendingMetadata, Password []byte) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	if meta.RemoteAddress.IP.String() != "127.0.0.1" {
		return false, meta.AuthFailed(), fmt.Errorf("invalid IP: %s", meta.RemoteAddress.IP.String())
	}
	if meta.ConnectionID != "0123456789ABCDEF" {
		return false, meta.AuthFailed(), fmt.Errorf("invalid connection ID: %s", meta.ConnectionID)
	}
	if meta.Username == "foo" && string(Password) == "bar" {
		return true, meta.AuthFailed(), nil
	}
	if meta.Username == "crash" {
		// Simulate a database failure
		return false, meta.AuthFailed(), fmt.Errorf("database error")
	}
	return false, meta.AuthFailed(), nil
}

func (h *handler) OnPubKey(meta metadata.ConnectionAuthPendingMetadata, publicKey auth3.PublicKey) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	if meta.RemoteAddress.IP.String() != "127.0.0.1" {
		return false, meta.AuthFailed(), fmt.Errorf("invalid IP: %s", meta.RemoteAddress.IP.String())
	}
	if meta.ConnectionID != "0123456789ABCDEF" {
		return false, meta.AuthFailed(), fmt.Errorf("invalid connection ID: %s", meta.ConnectionID)
	}
	if meta.Username == "foo" && publicKey.PublicKey == "ssh-rsa asdf" {
		return true, meta.AuthFailed(), nil
	}
	if meta.Username == "crash" {
		// Simulate a database failure
		return false, meta.AuthFailed(), fmt.Errorf("database error")
	}
	return false, meta.AuthFailed(), nil
}

func (h *handler) OnAuthorization(meta metadata.ConnectionAuthenticatedMetadata) (
	bool,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return false, meta.AuthFailed(), nil
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
		t.Run(
			fmt.Sprintf("subpath_%s", name), func(t *testing.T) {
				client, lifecycle, metricsCollector, err := initializeAuth(t, logger, subpath)
				if err != nil {
					assert.Fail(t, "failed to initialize auth", err)
					return
				}
				defer lifecycle.Stop(context.Background())

				authenticationContext := client.Password(
					metadata.NewTestAuthenticatingMetadata("foo"),
					[]byte("bar"),
				)
				assert.Equal(t, nil, authenticationContext.Error())
				assert.Equal(t, true, authenticationContext.Success())
				assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthBackendRequests)[0].Value)
				assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthSuccess)[0].Value)

				authenticationContext = client.Password(
					metadata.NewTestAuthenticatingMetadata("foo"),
					[]byte("baz"),
				)
				assert.Equal(t, nil, authenticationContext.Error())
				assert.Equal(t, false, authenticationContext.Success())
				assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthFailure)[0].Value)

				authenticationContext = client.Password(
					metadata.NewTestAuthenticatingMetadata("crash"),
					[]byte("baz"),
				)
				assert.NotEqual(t, nil, authenticationContext.Error())
				assert.Equal(t, false, authenticationContext.Success())
				assert.Equal(t, float64(1), metricsCollector.GetMetric(auth.MetricNameAuthBackendFailure)[0].Value)

				authenticationContext = client.PubKey(
					metadata.NewTestAuthenticatingMetadata("foo"),
					auth3.PublicKey{
						PublicKey: "ssh-rsa asdf",
					},
				)
				assert.Equal(t, nil, authenticationContext.Error())
				assert.Equal(t, true, authenticationContext.Success())

				authenticationContext = client.PubKey(
					metadata.NewTestAuthenticatingMetadata("foo"),
					auth3.PublicKey{
						PublicKey: "ssh-rsa asdx",
					},
				)
				assert.Equal(t, nil, authenticationContext.Error())
				assert.Equal(t, false, authenticationContext.Success())

				authenticationContext = client.PubKey(
					metadata.NewTestAuthenticatingMetadata("crash"),
					auth3.PublicKey{
						PublicKey: "ssh-rsa asdx",
					},
				)
				assert.NotEqual(t, nil, authenticationContext.Error())
				assert.Equal(t, false, authenticationContext.Success())
			},
		)
	}
}

func initializeAuth(t *testing.T, logger log.Logger, subpath string) (
	auth.WebhookClient,
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

	client, err := auth.NewWebhookClient(
		auth.AuthenticationTypeAll,
		config.AuthWebhookClientConfig{
			HTTPClientConfiguration: config.HTTPClientConfiguration{
				URL:     fmt.Sprintf("http://127.0.0.1:%d%s", port, subpath),
				Timeout: 2 * time.Second,
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
