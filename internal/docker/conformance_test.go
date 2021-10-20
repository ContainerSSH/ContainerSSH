package docker_test

import (
	"net"
	"os"
	"testing"

	"github.com/containerssh/geoip"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/metrics"
	sshserver "github.com/containerssh/sshserver/v2"
	"github.com/containerssh/structutils"
	"gopkg.in/yaml.v3"

	"github.com/containerssh/docker"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func(logger log.Logger) (sshserver.NetworkConnectionHandler, error){
		"dockerrun": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			//goland:noinspection GoDeprecation
			config := docker.DockerRunConfig{}
			structutils.Defaults(&config)

			testFile, err := os.Open("testdata/config-0.3.yaml")
			if err != nil {
				return nil, err
			}
			unmarshaller := yaml.NewDecoder(testFile)
			unmarshaller.KnownFields(true)
			if err := unmarshaller.Decode(&config); err != nil {
				return nil, err
			}

			return getDockerRun(config, logger)
		},
		"session": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			config := docker.Config{}
			structutils.Defaults(&config)

			config.Execution.Mode = docker.ExecutionModeSession
			return getDocker(config, logger)
		},
		"connection": func(logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
			config := docker.Config{}
			structutils.Defaults(&config)

			config.Execution.Mode = docker.ExecutionModeConnection
			return getDocker(config, logger)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}

func getDocker(config docker.Config, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	connectionID := sshserver.GenerateConnectionID()
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		return nil, err
	}
	collector := metrics.New(geoipProvider)
	return docker.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		connectionID,
		config,
		logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}

//goland:noinspection GoDeprecation
func getDockerRun(config docker.DockerRunConfig, logger log.Logger) (sshserver.NetworkConnectionHandler, error) {
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		return nil, err
	}
	collector := metrics.New(geoipProvider)
	return docker.NewDockerRun(
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
