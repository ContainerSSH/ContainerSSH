package kubernetes_test

import (
	"net"
	"testing"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/geoip/dummy"
	"github.com/containerssh/libcontainerssh/internal/kubernetes"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/internal/test"
	"github.com/containerssh/libcontainerssh/log"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"session": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg, err := getKubernetesConfig()
			if err != nil {
				return nil, err
			}
			cfg.Pod.Mode = config.KubernetesExecutionModeSession
			return getKubernetes(t, cfg, logger)
		},
		"connection": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg, err := getKubernetesConfig()
			if err != nil {
				return nil, err
			}
			cfg.Pod.Mode = config.KubernetesExecutionModeConnection
			return getKubernetes(t, cfg, logger)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}

func getKubernetes(t *testing.T, cfg config.KubernetesConfig, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	connectionID := sshserver.GenerateConnectionID()
	collector := metrics.New(dummy.New())
	return kubernetes.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: test.GetNextPort(t),
			Zone: "",
		}, connectionID, cfg, logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}

func getKubernetesConfig() (config.KubernetesConfig, error) {
	cfg := config.KubernetesConfig{}
	structutils.Defaults(&cfg)
	err := cfg.SetConfigFromKubeConfig()
	return cfg, err
}
