package config_test

import (
	"encoding/json"
	"testing"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/test"
	"gopkg.in/yaml.v3"
)

// issue639YAML is the docker section of the config.yaml from issue #639: the networking config is placed
// inside the "container" section, following the shape of the Docker container create request body.
const issue639YAML = `
backend: docker
docker:
  connection:
    host: "unix:///var/run/docker.sock"
  execution:
    container:
      image: custom-agent/containerssh-agent:v0
      networkingconfig:
        endpointsconfig:
          administrator_default:
            networkid: "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1"
    network:
      name: "administrator_default"
    mode: session
`

func networkIDOf(t *testing.T, cfg config.AppConfig, endpointName string) (string, bool) {
	t.Helper()
	networkConfig := cfg.Docker.Execution.NetworkConfig
	if networkConfig == nil || networkConfig.EndpointsConfig == nil {
		return "", false
	}
	endpoint, ok := networkConfig.EndpointsConfig[endpointName]
	if !ok || endpoint == nil {
		return "", false
	}
	return endpoint.NetworkID, true
}

// TestIssue639NetworkingConfigUnderContainerYAML ensures that a networkingconfig placed inside the
// "container" section (as in the Docker create request body) is picked up, see issue #639.
func TestIssue639NetworkingConfigUnderContainerYAML(t *testing.T) {
	cfg := config.AppConfig{}
	test.AssertNoError(t, yaml.Unmarshal([]byte(issue639YAML), &cfg))
	networkID, ok := networkIDOf(t, cfg, "administrator_default")
	test.AssertEquals(t, ok, true)
	test.AssertEquals(t, networkID, "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1")
}

// TestIssue639NetworkingConfigUnderContainerJSON is the JSON variant of the issue #639 config, using the
// Docker API style capitalized field names.
func TestIssue639NetworkingConfigUnderContainerJSON(t *testing.T) {
	data := `{
  "backend": "docker",
  "docker": {
    "connection": {"host": "unix:///var/run/docker.sock"},
    "execution": {
      "container": {
        "image": "custom-agent/containerssh-agent:v0",
        "NetworkingConfig": {
          "EndpointsConfig": {
            "administrator_default": {
              "NetworkID": "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1"
            }
          }
        }
      },
      "mode": "session"
    }
  }
}`
	cfg := config.AppConfig{}
	test.AssertNoError(t, json.Unmarshal([]byte(data), &cfg))
	networkID, ok := networkIDOf(t, cfg, "administrator_default")
	test.AssertEquals(t, ok, true)
	test.AssertEquals(t, networkID, "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1")
}

// TestIssue639NetworkingConfigUnderContainerYAMLCapitalized ensures the Docker API style capitalized keys
// also work in YAML documents.
func TestIssue639NetworkingConfigUnderContainerYAMLCapitalized(t *testing.T) {
	data := `
backend: docker
docker:
  execution:
    container:
      image: custom-agent/containerssh-agent:v0
      NetworkingConfig:
        EndpointsConfig:
          administrator_default:
            NetworkID: "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1"
`
	cfg := config.AppConfig{}
	test.AssertNoError(t, yaml.Unmarshal([]byte(data), &cfg))
	networkID, ok := networkIDOf(t, cfg, "administrator_default")
	test.AssertEquals(t, ok, true)
	test.AssertEquals(t, networkID, "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1")
}

// TestDockerLaunchConfigNetworkingConfigMerged ensures that a networking config placed under the documented
// "network" key and one placed inside "container" are merged, with the "network" key taking precedence on
// conflicts.
func TestDockerLaunchConfigNetworkingConfigMerged(t *testing.T) {
	data := `
backend: docker
docker:
  execution:
    container:
      image: custom-agent/containerssh-agent:v0
      networkingconfig:
        endpointsconfig:
          from_container:
            networkid: "aaaa"
          shared:
            networkid: "from_container"
    network:
      endpointsconfig:
        from_network:
          networkid: "bbbb"
        shared:
          networkid: "from_network"
`
	cfg := config.AppConfig{}
	test.AssertNoError(t, yaml.Unmarshal([]byte(data), &cfg))
	for endpoint, expected := range map[string]string{
		"from_container": "aaaa",
		"from_network":   "bbbb",
		"shared":         "from_network",
	} {
		networkID, ok := networkIDOf(t, cfg, endpoint)
		test.AssertEquals(t, ok, true)
		test.AssertEquals(t, networkID, expected)
	}
}

// TestDockerLaunchConfigNoNetworkingConfig ensures the default behavior is preserved: when no networking
// config is set anywhere, NetworkConfig stays nil so the Docker daemon applies its defaults.
func TestDockerLaunchConfigNoNetworkingConfig(t *testing.T) {
	data := `
backend: docker
docker:
  execution:
    container:
      image: custom-agent/containerssh-agent:v0
`
	cfg := config.AppConfig{}
	test.AssertNoError(t, yaml.Unmarshal([]byte(data), &cfg))
	if cfg.Docker.Execution.NetworkConfig != nil {
		t.Errorf(
			"NetworkConfig should be nil when not configured, got %+v",
			cfg.Docker.Execution.NetworkConfig,
		)
	}
}

// TestDockerLaunchConfigNetworkKeyStillWorks ensures the documented "network" key keeps working on its own.
func TestDockerLaunchConfigNetworkKeyStillWorks(t *testing.T) {
	data := `
backend: docker
docker:
  execution:
    container:
      image: custom-agent/containerssh-agent:v0
    network:
      endpointsconfig:
        administrator_default:
          networkid: "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1"
`
	cfg := config.AppConfig{}
	test.AssertNoError(t, yaml.Unmarshal([]byte(data), &cfg))
	networkID, ok := networkIDOf(t, cfg, "administrator_default")
	test.AssertEquals(t, ok, true)
	test.AssertEquals(t, networkID, "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1")
}
