package config_test

import (
	"context"
	"errors"
	"os"
	"testing"

    configuration "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/config"
    "go.containerssh.io/libcontainerssh/internal/structutils"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
	"github.com/stretchr/testify/assert"
)

// Test04Compatibility tests if a configuration file for ContainerSSH version 0.4 can be read.
func Test04Compatibility(t *testing.T) {
	logger := log.NewTestLogger(t)

	logger.Info(message.NewMessage("TEST", "FYI: any deprecation notices in this test are intentional."))

	testFile, err := os.Open("_testdata/0.4.yaml")
	assert.NoError(t, err)
	reader, err := config.NewReaderLoader(
		testFile,
		logger,
		config.FormatYAML,
	)
	assert.NoError(t, err)

	cfg := configuration.AppConfig{}
	structutils.Defaults(&cfg)
	err = reader.Load(context.Background(), &cfg)
	assert.NoError(t, err)
	err = cfg.Validate(false)
	assert.NoError(t, err)
}

// TestDockerRunRemoval tests if the removed DockerRun backend returns the correct error message.
func TestDockerRunRemoval(t *testing.T) {
	logger := log.NewTestLogger(t)

	testFile, err := os.Open("_testdata/dockerrun.yaml")
	assert.NoError(t, err)
	reader, err := config.NewReaderLoader(
		testFile,
		logger,
		config.FormatYAML,
	)
	assert.NoError(t, err)
	cfg := configuration.AppConfig{}
	structutils.Defaults(&cfg)
	err = reader.Load(context.Background(), &cfg)
	assert.NoError(t, err)
	err = cfg.Validate(false)
	if err == nil {
		t.Fatalf("Validation of a configuration with the DockerRun backend did not fail.")
	}
	var typedError message.Message
	if !errors.As(err, &typedError) {
		t.Fatalf("Using the deprecated DockerRun backend did not result in a message.Message error.")
	}
	if typedError.Code() != message.EDockerRunRemoved {
		t.Fatalf("Validation returned the wrong error code: %s", typedError.Code())
	}
}

// TestKubeRunRemoval tests if the removed KubeRun backend returns the correct error message.
func TestKubeRunRemoval(t *testing.T) {
	logger := log.NewTestLogger(t)

	testFile, err := os.Open("_testdata/kuberun.yaml")
	assert.NoError(t, err)
	reader, err := config.NewReaderLoader(
		testFile,
		logger,
		config.FormatYAML,
	)
	assert.NoError(t, err)
	cfg := configuration.AppConfig{}
	structutils.Defaults(&cfg)
	err = reader.Load(context.Background(), &cfg)
	assert.NoError(t, err)
	err = cfg.Validate(false)
	if err == nil {
		t.Fatalf("Validation of a configuration with the KubeRun backend did not fail.")
	}
	var typedError message.Message
	if !errors.As(err, &typedError) {
		t.Fatalf("Using the deprecated KubeRun backend did not result in a message.Message error.")
	}
	if typedError.Code() != message.EKubeRunRemoved {
		t.Fatalf("Validation returned the wrong error code: %s", typedError.Code())
	}
}
