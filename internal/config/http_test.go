package config_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	configuration "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/config"
	"github.com/containerssh/libcontainerssh/internal/geoip"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
	service2 "github.com/containerssh/libcontainerssh/service"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
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
		})
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

	cfg, err := client.Get(
		context.Background(),
		"foo",
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
		},
		connectionID,
		nil,
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
