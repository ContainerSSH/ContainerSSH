package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/structutils"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/metadata"
	"gopkg.in/yaml.v3"
)

// NewReaderLoader loads YAML files from reader.
func NewReaderLoader(
	reader io.Reader,
	logger log.Logger,
	format Format,
) (Loader, error) {
	if err := format.Validate(); err != nil {
		return nil, err
	}
	return &readerLoader{
		reader: reader,
		logger: logger,
		format: format,
	}, nil
}

type readerLoader struct {
	reader io.Reader
	logger log.Logger
	format Format
}

func (y *readerLoader) Load(_ context.Context, config *config.AppConfig) (err error) {
	switch y.format {
	case FormatYAML:
		err = y.loadYAML(y.reader, config)
	case FormatJSON:
		err = y.loadJSON(y.reader, config)
	default:
		err = fmt.Errorf("invalid format: %s", y.format)
	}
	if err != nil {
		return err
	}
	if err := fixCompatibility(config, y.logger); err != nil {
		return err
	}
	structutils.Defaults(config)
	return nil
}

func (y *readerLoader) LoadConnection(
	_ context.Context,
	meta metadata.ConnectionAuthenticatedMetadata,
	_ *config.AppConfig,
) (metadata.ConnectionAuthenticatedMetadata, error) {
	return meta, nil
}

func (y *readerLoader) loadYAML(reader io.Reader, config *config.AppConfig) error {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	return decoder.Decode(config)
}

func (y *readerLoader) loadJSON(reader io.Reader, config *config.AppConfig) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(config)
}
