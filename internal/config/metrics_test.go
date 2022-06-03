package config_test

import (
	"testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
)

func TestListenDefault(t *testing.T) {
	cfg := config.MetricsConfig{}
	structutils.Defaults(&cfg)
	assert.Equal(t, "0.0.0.0:9100", cfg.Listen)
}
