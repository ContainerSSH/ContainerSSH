package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/containerssh/containerssh/config"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"strings"
)

func parseConfig(filename string, data []byte, config *config.AppConfig) error {
	var err error
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		decoder := yaml.NewDecoder(bytes.NewReader(data))
		decoder.KnownFields(true)
		err = decoder.Decode(config)
		if err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, ".json") {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(config)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown file extension (%s)", filename)
	}
	return nil
}

func Write(config *config.AppConfig, writer io.Writer) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func LoadFile(filename string) (*config.AppConfig, error) {
	appConfig := &config.AppConfig{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = parseConfig(filename, data, appConfig)
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}
