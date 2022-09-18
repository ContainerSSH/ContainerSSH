package config_test

import (
    "encoding/json"
    "testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/test"
)

func TestIssue526UnknownFieldContainer(t *testing.T) {
    data := `{
"backend": "docker",
  "docker": {
    "execution": {
      "disableAgent": true,
      "idleCommand": [
        "/bin/sh",
        "-c",
        "sleep infinity & PID=$!; trap \\\"kill $PID\\\" INT TERM; wait"
      ],
      "container": {
        "image": "some_image"
      }
    }
  }
}`
    cfg := config.AppConfig{}
    test.AssertNoError(t, json.Unmarshal([]byte(data), &cfg))
    test.AssertNotNil(t, cfg.Docker.Execution.DockerLaunchConfig.ContainerConfig)
    test.AssertEquals(t, cfg.Docker.Execution.DockerLaunchConfig.ContainerConfig.Image, "some_image")
}
