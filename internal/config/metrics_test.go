package config_test

import (
	"testing"

    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
)

func TestListenDefault(t *testing.T) {
	cfg := config.MetricsConfig{}
	structutils.Defaults(&cfg)
	assert.Equal(t, "0.0.0.0:9100", cfg.Listen)
}
