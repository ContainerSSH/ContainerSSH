package docker_test

import (
	"net"
	"testing"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/docker"
	"github.com/containerssh/containerssh/internal/geoip/dummy"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"session": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := config.DockerConfig{}
			structutils.Defaults(&cfg)

			cfg.Execution.Mode = config.DockerExecutionModeSession
			return getDocker(cfg, logger)
		},
		"connection": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := config.DockerConfig{}
			structutils.Defaults(&cfg)

			cfg.Execution.Mode = config.DockerExecutionModeConnection
			return getDocker(cfg, logger)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}

func getDocker(cfg config.DockerConfig, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	connectionID := sshserver.GenerateConnectionID()
	collector := metrics.New(dummy.New())
	return docker.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		connectionID,
		cfg,
		logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}
