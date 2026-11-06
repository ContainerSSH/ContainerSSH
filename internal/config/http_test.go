package config_test

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"

    configuration "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/config"
    "go.containerssh.io/containerssh/internal/geoip"
    "go.containerssh.io/containerssh/internal/metrics"
    "go.containerssh.io/containerssh/internal/test"
    "go.containerssh.io/containerssh/log"
    "go.containerssh.io/containerssh/metadata"
    service2 "go.containerssh.io/containerssh/service"
)

func TestHTTP(t *testing.T) {
	port := test.GetNextPort(t, "HTTP")
	logger := log.NewTestLogger(t)
	srv, err := config.NewServer(
		configuration.HTTPServerConfiguration{
			Listen: fmt.Sprintf("127.0.0.1:%d", port),
		},
		&myConfigReqHandler{},
		logger,
	)
	assert.NoError(t, err)
	lifecycle := service2.NewLifecycle(srv)

	ready := make(chan struct{})
	lifecycle.OnRunning(
		func(s service2.Service, l service2.Lifecycle) {
			ready <- struct{}{}
		},
	)
	go func() {
		_ = lifecycle.Run()
	}()
	<-ready

	client, err := config.NewClient(
		configuration.ClientConfig{
			HTTPClientConfiguration: configuration.HTTPClientConfiguration{
				URL:     fmt.Sprintf("http://127.0.0.1:%d", port),
				Timeout: 2 * time.Second,
			},
		}, logger, getMetricsCollector(t),
	)
	assert.NoError(t, err)

	connectionID := "0123456789ABCDEF"

	cfg, _, err := client.Get(
		context.Background(),
		metadata.ConnectionAuthenticatedMetadata{
			ConnectionAuthPendingMetadata: metadata.ConnectionAuthPendingMetadata{
				ConnectionMetadata: metadata.ConnectionMetadata{
					RemoteAddress: metadata.RemoteAddress(
						net.TCPAddr{
							IP:   net.ParseIP("127.0.0.1"),
							Port: port,
						},
					),
					ConnectionID: connectionID,
					Metadata:     map[string]metadata.Value{},
					Environment:  map[string]metadata.Value{},
					Files:        map[string]metadata.BinaryValue{},
				},
				ClientVersion: "",
				Username:      "foo",
			},
			AuthenticatedUsername: "foo",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, "yourcompany/yourimage", cfg.Docker.Execution.DockerLaunchConfig.ContainerConfig.Image)

	lifecycle.Stop(context.Background())
	err = lifecycle.Wait()
	assert.NoError(t, err)
}

func TestHTTPRetryIsLoggedAsWarning(t *testing.T) {
	logOutput := &bytes.Buffer{}
	logger := log.MustNewLogger(configuration.LogConfig{
		Level:       configuration.LogLevelInfo,
		Format:      configuration.LogFormatLJSON,
		Destination: configuration.LogDestinationStdout,
		Stdout:      logOutput,
	})
	client, err := config.NewClient(
		configuration.ClientConfig{
			HTTPClientConfiguration: configuration.HTTPClientConfiguration{
				// Nothing listens on this port, so every request fails.
				URL:     fmt.Sprintf("http://127.0.0.1:%d", test.GetNextPort(t, "config server")),
				Timeout: time.Second,
			},
		}, logger, getMetricsCollector(t),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, err = client.Get(ctx, metadata.NewTestAuthenticatingMetadata("foo").Authenticated("foo"))
	assert.Error(t, err)
	assert.Contains(t, logOutput.String(), `"level":"warning","code":"CONFIG_BACKEND_ERROR","message":"Configuration request to backend failed, retrying`)
}

func getMetricsCollector(t *testing.T) metrics.Collector {
	geoIP, err := geoip.New(configuration.GeoIPConfig{
		Provider: "dummy",
	})
	assert.NoError(t, err)
	return metrics.New(geoIP)
}

type myConfigReqHandler struct {
}

func (m *myConfigReqHandler) OnConfig(
	request configuration.Request,
) (config configuration.AppConfig, err error) {
	config.Backend = "docker"
	config.Docker.Execution.DockerLaunchConfig.ContainerConfig = &container.Config{}
	if request.Username == "foo" {
		config.Docker.Execution.DockerLaunchConfig.ContainerConfig.Image = "yourcompany/yourimage"
	}
	return config, err
}
