package config

import (
	"encoding/json"
	"fmt"
	"io"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
	"gopkg.in/yaml.v3"
)

// NewWriterSaver creates a config saver that writes the testdata in YAML format to the specified writer.
func NewWriterSaver(
	writer io.Writer,
	logger log.Logger,
	format Format,
) (ConfigSaver, error) {
	return &writerSaver{
		writer: writer,
		logger: logger,
		format: format,
	}, nil
}

type writerSaver struct {
	writer io.Writer
	logger log.Logger
	format Format
}

func (w *writerSaver) Save(config *config.AppConfig) error {
	switch w.format {
	case FormatYAML:
		return w.saveYAML(config)
	case FormatJSON:
		return w.saveJSON(config)
	default:
		return fmt.Errorf("invalid format: %s", w.format)
	}
}

func (w *writerSaver) saveYAML(config *config.AppConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (w *writerSaver) saveJSON(config *config.AppConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}
