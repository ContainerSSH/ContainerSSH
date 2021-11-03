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

// Test04Compatibility tests if a configuration file for ContainerSSH version 0.4 can be read.
func Test04Compatibility(t *testing.T) {
	t.Skipf("This test hasn't been implemented yet.")
	logger := log.NewTestLogger(t)

	logger.Info(message.NewMessage("TEST", "FYI: the deprecation notice in this test is intentional"))

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
}
