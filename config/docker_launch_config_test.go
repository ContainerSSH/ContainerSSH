package config_test

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"go.containerssh.io/containerssh/config"
	"gopkg.in/yaml.v3"
)

func TestDockerLaunchConfigValidateNetworkWithoutEndpoints(t *testing.T) {
	cfgText := "container:\n  image: \"containerssh/containerssh-guest-image\"\nnetwork:\n  name: \"administrator_default\"\n"
	cfg := config.DockerLaunchConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	err := cfg.Validate()
	assert.Error(
		t,
		err,
		"a network block that decodes to no endpoints (e.g. from a misspelled field such as \"name\" "+
			"instead of \"endpointsConfig\") should be rejected instead of silently falling back to the default network",
	)
}

func TestDockerLaunchConfigValidateNetworkWithEndpoints(t *testing.T) {
	cfgText := "container:\n  image: \"containerssh/containerssh-guest-image\"\n" +
		"network:\n  endpointsconfig:\n    administrator_default:\n      networkid: \"abc123\"\n"
	cfg := config.DockerLaunchConfig{}
	assert.NoError(t, yaml.Unmarshal([]byte(cfgText), &cfg))

	assert.NoError(t, cfg.Validate())
}

func TestDockerLaunchConfigValidateNoNetwork(t *testing.T) {
	cfg := config.DockerLaunchConfig{
		ContainerConfig: &container.Config{
			Image: "containerssh/containerssh-guest-image",
		},
	}

	assert.NoError(t, cfg.Validate())
}
