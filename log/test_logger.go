package log

import (
	"testing"

    "go.containerssh.io/libcontainerssh/config"
)

// NewTestLogger creates a logger for testing purposes.
//goland:noinspection GoUnusedExportedFunction
func NewTestLogger(t *testing.T) Logger {
	logger, err := NewLogger(
		config.LogConfig{
			Level:       config.LogLevelDebug,
			Format:      config.LogFormatText,
			Destination: config.LogDestinationTest,
			T:           t,
		},
	)
	if err != nil {
		panic(err)
	}
	return logger
}
