package config_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/registry"
	"go.containerssh.io/containerssh/config"
	"gopkg.in/yaml.v3"
)

// dockerExecutionConfigFixture returns a DockerExecutionConfig with every field set to a non-zero, non-default value.
func dockerExecutionConfigFixture() config.DockerExecutionConfig {
	return config.DockerExecutionConfig{
		DockerLaunchConfig: config.DockerLaunchConfig{
			ContainerConfig: &container.Config{Image: "registry.example.com/private/image:1.0"},
		},
		Auth: &registry.AuthConfig{
			Username:      "user",
			Password:      "secret",
			ServerAddress: "registry.example.com",
		},
		Mode:                    config.DockerExecutionModeSession,
		IdleCommand:             []string{"/bin/sleep", "infinity"},
		ShellCommand:            []string{"/bin/zsh"},
		AgentPath:               "/opt/agent",
		DisableAgent:            true,
		Subsystems:              map[string]string{"sftp": "/opt/sftp-server"},
		ImagePullPolicy:         config.ImagePullPolicyNever,
		ExposeAuthMetadataAsEnv: true,
	}
}

// TestDockerExecutionConfigRoundTrip checks that every field of DockerExecutionConfig survives both custom
// unmarshallers. UnmarshalJSON and UnmarshalYAML decode into a shadow struct and copy fields over by hand, so a
// field missing from that copy block is silently dropped (see issue #707 for docker.execution.auth).
func TestDockerExecutionConfigRoundTrip(t *testing.T) {
	expected := dockerExecutionConfigFixture()

	// Guard: a newly added field must be populated in the fixture, otherwise it would not be covered below.
	v := reflect.ValueOf(expected)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			t.Fatalf("fixture does not set DockerExecutionConfig.%s, please extend it", v.Type().Field(i).Name)
		}
	}

	formats := map[string]struct {
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		"json": {json.Marshal, json.Unmarshal},
		"yaml": {yaml.Marshal, yaml.Unmarshal},
	}

	for name, format := range formats {
		t.Run(name, func(t *testing.T) {
			data, err := format.marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			var actual config.DockerExecutionConfig
			if err := format.unmarshal(data, &actual); err != nil {
				t.Fatal(err)
			}

			e := reflect.ValueOf(expected)
			a := reflect.ValueOf(actual)
			for i := 0; i < e.NumField(); i++ {
				field := e.Type().Field(i).Name
				if field == "DockerLaunchConfig" {
					// Covered by DockerLaunchConfig's own unmarshaller; only check it was carried over at all.
					if actual.ContainerConfig == nil || actual.ContainerConfig.Image != expected.ContainerConfig.Image {
						t.Errorf("DockerLaunchConfig was not preserved: got %+v", actual.ContainerConfig)
					}
					continue
				}
				if !reflect.DeepEqual(e.Field(i).Interface(), a.Field(i).Interface()) {
					t.Errorf(
						"field %s was dropped or altered: expected %+v, got %+v",
						field,
						e.Field(i).Interface(),
						a.Field(i).Interface(),
					)
				}
			}
		})
	}
}
