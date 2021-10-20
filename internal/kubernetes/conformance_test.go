package kubernetes_test

import (
	"net"
	"os"
	"testing"

	"github.com/containerssh/geoip"
	"github.com/containerssh/kubernetes/v3"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/sshserver/v2"
	"github.com/containerssh/structutils"
	"gopkg.in/yaml.v3"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"session": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			config, err := getKubernetesConfig()
			if err != nil {
				return nil, err
			}
			config.Pod.Mode = kubernetes.ExecutionModeSession
			return getKubernetes(config, logger)
		},
		"connection": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			config, err := getKubernetesConfig()
			if err != nil {
				return nil, err
			}
			config.Pod.Mode = kubernetes.ExecutionModeConnection
			return getKubernetes(config, logger)
		},
		"kuberun": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			config, err := getKubeRunConfig()
			if err != nil {
				return nil, err
			}
			config.Pod.EnableAgent = true
			config.Pod.ShellCommand = []string{"/bin/bash"}
			return getKubeRun(config, logger)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}

func getKubernetes(config kubernetes.Config, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	connectionID := sshserver.GenerateConnectionID()
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		return nil, err
	}
	collector := metrics.New(geoipProvider)
	return kubernetes.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		}, connectionID, config, logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}

func getKubernetesConfig() (kubernetes.Config, error) {
	config := kubernetes.Config{}
	structutils.Defaults(&config)
	err := config.SetConfigFromKubeConfig()
	return config, err
}

//goland:noinspection GoDeprecation
func getKubeRunConfig() (kubernetes.KubeRunConfig, error) {
	fh, err := os.Open("testdata/config-0.3.yaml")
	if err != nil {
		return kubernetes.KubeRunConfig{}, err
	}
	defer func() {
		_ = fh.Close()
	}()

	//goland:noinspection GoDeprecation
	config := kubernetes.KubeRunConfig{}
	structutils.Defaults(&config)
	fileConfig := kubernetes.KubeRunConfig{}
	decoder := yaml.NewDecoder(fh)
	decoder.KnownFields(true)
	if err := decoder.Decode(&fileConfig); err != nil {
		return config, err
	}
	if err := structutils.Merge(&config, &fileConfig); err != nil {
		return config, err
	}
	if err := config.SetConfigFromKubeConfig(); err != nil {
		return config, err
	}
	return config, nil
}

//goland:noinspection GoDeprecation
func getKubeRun(config kubernetes.KubeRunConfig, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		return nil, err
	}
	collector := metrics.New(geoipProvider)
	//goland:noinspection GoDeprecation
	return kubernetes.NewKubeRun(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		sshserver.GenerateConnectionID(),
		config,
		logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}
