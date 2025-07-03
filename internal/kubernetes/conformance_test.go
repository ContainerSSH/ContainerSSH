package kubernetes_test

import (
	"net"
	"testing"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/geoip/dummy"
	"go.containerssh.io/containerssh/internal/kubernetes"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/internal/sshserver"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/internal/test"
	"go.containerssh.io/containerssh/log"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"session": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := getKubernetesConfig(t)
			cfg.Pod.Mode = config.KubernetesExecutionModeSession
			return getKubernetes(t, cfg, logger)
		},
		"connection": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := getKubernetesConfig(t)
			cfg.Pod.Mode = config.KubernetesExecutionModeConnection
			return getKubernetes(t, cfg, logger)
		},
		"persistent": func(t *testing.T, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			cfg := getKubernetesConfig(t)
			cfg.Pod.Mode = config.KubernetesExecutionModePersistent
			cfg.Pod.CreateMissingPods = true
			cfg.Pod.Metadata.Name = "containerssh-test-pod"
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
			Port: test.GetNextPort(t, "client"),
			Zone: "",
		}, connectionID, cfg, logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}

func getKubernetesConfig(t *testing.T) config.KubernetesConfig {
	cfg := config.KubernetesConfig{}
	structutils.Defaults(&cfg)
	kube := test.Kubernetes(t)
	cfg.Connection.Host = kube.Host
	cfg.Connection.CAData = kube.CACert
	cfg.Connection.CertData = kube.UserCert
	cfg.Connection.KeyData = kube.UserKey
	cfg.Connection.ServerName = kube.ServerName
	return cfg
}
