package log

import (
	"io"
	"os"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/message"
)

// NewLogger creates a standard logger pipeline.
//
// - config is the configuration structure for the logger
//
// goland:noinspection GoUnusedExportedFunction
func NewLogger(config config.LogConfig) (Logger, error) {
	return NewLoggerFactory().Make(config)
}

// MustNewLogger is identical to NewLogger, except that it panics instead of returning an error
func MustNewLogger(
	config config.LogConfig,
) Logger {
	logger, err := NewLogger(config)
	if err != nil {
		panic(err)
	}
	return logger
}

// NewLoggerFactory create a pipeline logger factory
//goland:noinspection GoUnusedExportedFunction
func NewLoggerFactory() LoggerFactory {
	return &loggerFactory{}
}

type loggerFactory struct {
}

func (f *loggerFactory) MustMake(config config.LogConfig) Logger {
	logger, err := f.Make(config)
	if err != nil {
		panic(err)
	}
	return logger
}

func (f *loggerFactory) Make(cfg config.LogConfig) (Logger, error) {
	if err := cfg.Level.Validate(); err != nil {
		return nil, err
	}

	if err := cfg.Format.Validate(); err != nil {
		return nil, err
	}

	if err := cfg.Destination.Validate(); err != nil {
		return nil, err
	}

	var writer Writer
	var err error = nil
	helper := func() {}
	switch cfg.Destination {
	case config.LogDestinationFile:
		writer, err = newFileWriter(cfg.File, cfg.Format)
	case config.LogDestinationStdout:
		var stdout io.Writer = os.Stdout
		if cfg.Stdout != nil {
			stdout = cfg.Stdout
		}
		writer = newStdoutWriter(stdout, cfg.Format)
	case config.LogDestinationSyslog:
		writer, err = newSyslogWriter(cfg.Syslog, cfg.Format)
	case config.LogDestinationTest:
		writer = newGoTest(cfg.T)
		helper = cfg.T.Helper
	}
	if err != nil {
		return nil, err
	}

	return &logger{
		level:  cfg.Level,
		labels: map[message.LabelName]message.LabelValue{},
		writer: writer,
		helper: helper,
	}, nil
}
