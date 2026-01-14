package config_test

import (
	"context"
	"os"
	"testing"

    configuration "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/config"
    "go.containerssh.io/containerssh/internal/structutils"
    "go.containerssh.io/containerssh/log"
    "go.containerssh.io/containerssh/message"
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
