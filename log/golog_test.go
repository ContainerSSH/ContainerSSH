package log_test

import (
	"bytes"
	goLog "log"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
	"github.com/stretchr/testify/assert"
)

func TestGoLog(t *testing.T) {
	writer := &bytes.Buffer{}
	logger := log.MustNewLogger(
		config.LogConfig{
			Level:       config.LogLevelDebug,
			Format:      config.LogFormatText,
			Destination: config.LogDestinationStdout,
			Stdout:      writer,
		},
	)
	goLogWriter := log.NewGoLogWriter(logger)
	goLogger := goLog.New(goLogWriter, "", 0)
	goLogger.Printf("test")
	assert.True(t, len(writer.Bytes()) > 0)
}
