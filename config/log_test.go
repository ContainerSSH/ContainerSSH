package config_test

import (
	"encoding/json"
	"testing"

	"github.com/containerssh/structutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/containerssh/containerssh/log"
)

func TestJSONDecode(t *testing.T) {
	cfg := "{\"level\":\"debug\",\"format\":\"text\"}"
	config := log.Config{}
	err := json.Unmarshal([]byte(cfg), &config)
	assert.NoError(t, err)
	assert.Equal(t, log.LevelDebug, config.Level)
	assert.Equal(t, log.FormatText, config.Format)
}

func TestJSONDecodeNumeric(t *testing.T) {
	cfg := "{\"level\":7,\"format\":\"text\"}"
	config := log.Config{}
	err := json.Unmarshal([]byte(cfg), &config)
	assert.NoError(t, err)
	assert.Equal(t, log.LevelDebug, config.Level)
	assert.Equal(t, log.FormatText, config.Format)
}

func TestYAMLDecode(t *testing.T) {
	cfg := "---\nlevel: debug\nformat: text\n"
	config := log.Config{}
	err := yaml.Unmarshal([]byte(cfg), &config)
	assert.NoError(t, err)
	assert.Equal(t, log.LevelDebug, config.Level)
	assert.Equal(t, log.FormatText, config.Format)
}

func TestYAMLDecodeNumeric(t *testing.T) {
	cfg := "---\nlevel: 7\nformat: text\n"
	config := log.Config{}
	err := yaml.Unmarshal([]byte(cfg), &config)
	assert.NoError(t, err)
	assert.Equal(t, log.LevelDebug, config.Level)
	assert.Equal(t, log.FormatText, config.Format)
}

func TestJSONEncode(t *testing.T) {
	config := log.Config{
		Level:  log.LevelDebug,
		Format: log.FormatText,
	}
	jsonData, err := json.Marshal(config)
	assert.NoError(t, err)
	rawData := map[string]interface{}{}
	err = json.Unmarshal(jsonData, &rawData)
	assert.NoError(t, err)

	assert.Equal(t, "debug", rawData["level"])
	assert.Equal(t, "text", rawData["format"])
}

func TestYAMLEncode(t *testing.T) {
	config := log.Config{
		Level:  log.LevelDebug,
		Format: log.FormatText,
	}
	jsonData, err := yaml.Marshal(config)
	assert.NoError(t, err)
	rawData := map[string]interface{}{}
	err = yaml.Unmarshal(jsonData, &rawData)
	assert.NoError(t, err)

	assert.Equal(t, "debug", rawData["level"])
	assert.Equal(t, "text", rawData["format"])
}

func TestDefault(t *testing.T) {
	config := log.Config{}
	structutils.Defaults(&config)
	assert.Equal(t, log.LevelNotice, config.Level)
	assert.Equal(t, log.FormatLJSON, config.Format)
}
