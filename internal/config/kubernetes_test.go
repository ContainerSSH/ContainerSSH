package config_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/structutils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoadIssue209(t *testing.T) {
	testFile, err := os.Open("_testdata/issue-209.yaml")
	assert.NoError(t, err)
	cfg := config.KubernetesConfig{}
	unmarshaller := yaml.NewDecoder(testFile)
	unmarshaller.KnownFields(true)
	assert.NoError(t, unmarshaller.Decode(&cfg))

	assert.Equal(t, "/home/ubuntu", cfg.Pod.Spec.Volumes[0].HostPath.Path)
}

func TestLoadSave(t *testing.T) {
	oldConfig := &config.KubernetesConfig{}
	newConfig := &config.KubernetesConfig{}

	structutils.Defaults(oldConfig)

	data := &bytes.Buffer{}
	encoder := yaml.NewEncoder(data)
	if err := encoder.Encode(oldConfig); err != nil {
		t.Fatal(err)
	}
	decoder := yaml.NewDecoder(data)
	if err := decoder.Decode(newConfig); err != nil {
		t.Fatal(err)
	}

	diff := cmp.Diff(
		oldConfig,
		newConfig,
		cmp.AllowUnexported(config.KubernetesPodConfig{}),
		cmp.AllowUnexported(config.KubernetesConnectionConfig{}),
		cmpopts.EquateEmpty(),
	)
	if diff != "" {
		t.Fatal(fmt.Errorf("restored configuration is different from the saved config: %v", diff))
	}
}
