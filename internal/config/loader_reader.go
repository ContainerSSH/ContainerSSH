package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/log"
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
	return nil
}

func (y *readerLoader) LoadConnection(
	_ context.Context,
	_ string,
	_ net.TCPAddr,
	_ string,
	_ *auth.ConnectionMetadata,
	_ *config.AppConfig,
) error {
	return nil
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
