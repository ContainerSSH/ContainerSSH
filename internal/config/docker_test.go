package config_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// TestYAMLSerialization tests if the configuration structure can be serialized and then deserialized to/from YAML.
func TestYAMLSerialization(t *testing.T) {
	t.Parallel()

	// region Setup
	cfg := &config.DockerConfig{}
	newCfg := &config.DockerConfig{}
	structutils.Defaults(cfg)

	buf := &bytes.Buffer{}
	// endregion

	// region Save
	yamlEncoder := yaml.NewEncoder(buf)
	assert.NoError(t, yamlEncoder.Encode(cfg))
	// endregion

	// region Load
	yamlDecoder := yaml.NewDecoder(buf)
	yamlDecoder.KnownFields(true)
	assert.NoError(t, yamlDecoder.Decode(newCfg))
	// endregion

	// region Assert

	diff := cmp.Diff(
		cfg,
		newCfg,
		cmp.AllowUnexported(config.DockerExecutionConfig{}),
		cmpopts.EquateEmpty(),
	)
	assert.Empty(t, diff)
	// endregion
}

// TestJSONSerialization tests if the configuration structure can be serialized and then deserialized to/from JSON.
func TestJSONSerialization(t *testing.T) {
	t.Parallel()

	// region Setup
	cfg := &config.DockerConfig{}
	newCfg := &config.DockerConfig{}
	structutils.Defaults(cfg)

	buf := &bytes.Buffer{}
	// endregion

	// region Save
	jsonEncoder := json.NewEncoder(buf)
	assert.NoError(t, jsonEncoder.Encode(cfg))
	// endregion

	// region Load
	jsonDecoder := json.NewDecoder(buf)
	jsonDecoder.DisallowUnknownFields()
	assert.NoError(t, jsonDecoder.Decode(newCfg))
	// endregion

	// region Assert

	diff := cmp.Diff(
		cfg,
		newCfg,
		cmp.AllowUnexported(config.DockerExecutionConfig{}),
		cmpopts.EquateEmpty(),
	)
	assert.Empty(t, diff)
	// endregion
}

// TestUnmarshalYAML03 tests the ContainerSSH 0.3 compatibility. It checks if a YAML fragment from 0.3 can still be
// unmarshalled.
func TestUnmarshalYAML03(t *testing.T) {
	t.Parallel()

	testFile, err := os.Open("testdata/config-0.3.yaml")
	assert.NoError(t, err)
	unmarshaller := yaml.NewDecoder(testFile)
	unmarshaller.KnownFields(true)
	//goland:noinspection GoDeprecation
	cfg := config.DockerRunConfig{}
	assert.NoError(t, unmarshaller.Decode(&cfg))
	assert.Equal(t, false, cfg.Config.DisableCommand)
	assert.Equal(t, "/usr/lib/openssh/sftp-server", cfg.Config.Subsystems["sftp"])
	assert.Equal(t, 60*time.Second, cfg.Config.Timeout)
}

// TestUnmarshalYAML03 tests the ContainerSSH 0.3 compatibility. It checks if a JSON fragment from 0.3 can still be
// unmarshalled.
func TestUnmarshalJSON03(t *testing.T) {
	t.Parallel()

	testFile, err := os.Open("testdata/config-0.3.json")
	assert.NoError(t, err)
	unmarshaller := json.NewDecoder(testFile)
	unmarshaller.DisallowUnknownFields()
	//goland:noinspection GoDeprecation
	cfg := config.DockerRunConfig{}
	assert.NoError(t, unmarshaller.Decode(&cfg))
	assert.Equal(t, false, cfg.Config.DisableCommand)
	assert.Equal(t, "/usr/lib/openssh/sftp-server", cfg.Config.Subsystems["sftp"])
	assert.Equal(t, 60*time.Second, cfg.Config.Timeout)
}


