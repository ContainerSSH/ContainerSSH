package config_test

import (
	"encoding/json"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestJSONDecode(t *testing.T) {
	cfgText := "{\"level\":\"debug\",\"format\":\"text\"}"
	cfg := config.LogConfig{}
	err := json.Unmarshal([]byte(cfgText), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, config.LogLevelDebug, cfg.Level)
	assert.Equal(t, config.LogFormatText, cfg.Format)
}

func TestJSONDecodeNumeric(t *testing.T) {
	cfgText := "{\"level\":7,\"format\":\"text\"}"
	cfg := config.LogConfig{}
	err := json.Unmarshal([]byte(cfgText), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, config.LogLevelDebug, cfg.Level)
	assert.Equal(t, config.LogFormatText, cfg.Format)
}

func TestYAMLDecode(t *testing.T) {
	cfgText := "---\nlevel: debug\nformat: text\n"
	cfg := config.LogConfig{}
	err := yaml.Unmarshal([]byte(cfgText), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, config.LogLevelDebug, cfg.Level)
	assert.Equal(t, config.LogFormatText, cfg.Format)
}

func TestYAMLDecodeNumeric(t *testing.T) {
	cfgText := "---\nlevel: 7\nformat: text\n"
	cfg := config.LogConfig{}
	err := yaml.Unmarshal([]byte(cfgText), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, config.LogLevelDebug, cfg.Level)
	assert.Equal(t, config.LogFormatText, cfg.Format)
}

func TestJSONEncode(t *testing.T) {
	cfg := config.LogConfig{
		Level:  config.LogLevelDebug,
		Format: config.LogFormatText,
	}
	jsonData, err := json.Marshal(cfg)
	assert.NoError(t, err)
	rawData := map[string]interface{}{}
	err = json.Unmarshal(jsonData, &rawData)
	assert.NoError(t, err)

	assert.Equal(t, "debug", rawData["level"])
	assert.Equal(t, "text", rawData["format"])
}

func TestYAMLEncode(t *testing.T) {
	cfg := config.LogConfig{
		Level:  config.LogLevelDebug,
		Format: config.LogFormatText,
	}
	jsonData, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	rawData := map[string]interface{}{}
	err = yaml.Unmarshal(jsonData, &rawData)
	assert.NoError(t, err)

	assert.Equal(t, "debug", rawData["level"])
	assert.Equal(t, "text", rawData["format"])
}

func TestDefault(t *testing.T) {
	cfg := config.LogConfig{}
	structutils.Defaults(&cfg)
	assert.Equal(t, config.LogLevelNotice, cfg.Level)
	assert.Equal(t, config.LogFormatLJSON, cfg.Format)
}
