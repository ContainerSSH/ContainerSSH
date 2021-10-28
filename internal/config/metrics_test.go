package config_test

import (
	"testing"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
)

func TestListenDefault(t *testing.T) {
	cfg := config.MetricsConfig{}
	structutils.Defaults(&cfg)
	assert.Equal(t, "0.0.0.0:9100", cfg.Listen)
}
