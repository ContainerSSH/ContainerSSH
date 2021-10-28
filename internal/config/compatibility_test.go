package config_test

import (
	"context"
	"os"
	"testing"

	configuration "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/config"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/stretchr/testify/assert"
)

// Test03Compatibility tests if a configuration file for ContainerSSH version 0.3 can be read.
func Test03Compatibility(t *testing.T) {
	logger := log.NewTestLogger(t)

	logger.Info(message.NewMessage("TEST", "FYI: the deprecation notice in this test is intentional"))

	testFile, err := os.Open("testdata/0.3.yaml")
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

	assert.Equal(t, "0.0.0.0:2222", cfg.SSH.Listen)
}
