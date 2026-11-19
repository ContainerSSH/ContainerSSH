package docker_test

import (
	"net"
	"testing"

	"github.com/docker/docker/api/types/container"

    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/docker"
    "go.containerssh.io/containerssh/internal/geoip/dummy"
    "go.containerssh.io/containerssh/internal/metrics"
    "go.containerssh.io/containerssh/internal/sshserver"
    "go.containerssh.io/containerssh/internal/structutils"
    "go.containerssh.io/containerssh/internal/test"
    "go.containerssh.io/containerssh/log"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"session": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := config.DockerConfig{}
			structutils.Defaults(&cfg)

			cfg.Execution.Mode = config.DockerExecutionModeSession
			return getDocker(t, cfg, logger)
		},
		"session-log-driver-none": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := config.DockerConfig{}
			structutils.Defaults(&cfg)

			cfg.Execution.Mode = config.DockerExecutionModeSession
			// A log driver that cannot be read must not break attaching to the container.
			cfg.Execution.HostConfig = &container.HostConfig{
				LogConfig: container.LogConfig{Type: "none"},
			}
			return getDocker(t, cfg, logger)
		},
		"connection": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := config.DockerConfig{}
			structutils.Defaults(&cfg)

			cfg.Execution.Mode = config.DockerExecutionModeConnection
			return getDocker(t, cfg, logger)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}

func getDocker(t *testing.T, cfg config.DockerConfig, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	connectionID := sshserver.GenerateConnectionID()
	collector := metrics.New(dummy.New())
	return docker.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: test.GetNextPort(t, "client"),
			Zone: "",
		},
		connectionID,
		cfg,
		logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}
