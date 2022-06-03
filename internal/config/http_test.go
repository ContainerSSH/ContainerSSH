package config_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"

    configuration "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/config"
    "go.containerssh.io/libcontainerssh/internal/geoip"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/internal/test"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/metadata"
    service2 "go.containerssh.io/libcontainerssh/service"
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
	assert.Equal(t, "yourcompany/yourimage", cfg.Docker.Execution.Launch.ContainerConfig.Image)

	lifecycle.Stop(context.Background())
	err = lifecycle.Wait()
	assert.NoError(t, err)
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
	config.Docker.Execution.Launch.ContainerConfig = &container.Config{}
	if request.Username == "foo" {
		config.Docker.Execution.Launch.ContainerConfig.Image = "yourcompany/yourimage"
	}
	return config, err
}
