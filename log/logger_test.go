package log_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/message"
	"github.com/stretchr/testify/assert"

    "go.containerssh.io/libcontainerssh/log"
)

func TestLogLevelFiltering(t *testing.T) {
	for logLevelInt := 0; logLevelInt < 8; logLevelInt++ {
		t.Run(fmt.Sprintf("filter=%s", config.LogLevel(logLevelInt).MustName()), func(t *testing.T) {
			for writeLogLevelInt := 0; writeLogLevelInt < 8; writeLogLevelInt++ {
				logLevel := config.LogLevel(logLevelInt)
				writeLogLevel := config.LogLevel(writeLogLevelInt)
				t.Run(
					fmt.Sprintf("write=%s", config.LogLevel(writeLogLevelInt).MustName()),
					func(t *testing.T) {
						testLevel(t, logLevel, writeLogLevel)
					},
				)
			}
		})
	}
}

func testLevel(t *testing.T, logLevel config.LogLevel, writeLogLevel config.LogLevel) {
	var buf bytes.Buffer
	p := log.MustNewLogger(config.LogConfig{
		Level:       logLevel,
		Format:      config.LogFormatLJSON,
		Destination: config.LogDestinationStdout,
		Stdout:      &buf,
	})
	msg := message.UserMessage("E_TEST", "test", "test")
	switch writeLogLevel {
	case config.LogLevelDebug:
		p.Debug(msg)
	case config.LogLevelInfo:
		p.Info(msg)
	case config.LogLevelNotice:
		p.Notice(msg)
	case config.LogLevelWarning:
		p.Warning(msg)
	case config.LogLevelError:
		p.Error(msg)
	case config.LogLevelCritical:
		p.Critical(msg)
	case config.LogLevelAlert:
		p.Alert(msg)
	case config.LogLevelEmergency:
		p.Emergency(msg)
	}
	if logLevel < writeLogLevel {
		assert.Equal(t, 0, buf.Len())
	} else {
		assert.NotEqual(t, 0, buf.Len())

		rawData := buf.Bytes()
		data := map[string]interface{}{}
		if err := json.Unmarshal(rawData, &data); err != nil {
			assert.Fail(t, "failed to unmarshal JSON from writer", err)
		}

		expectedLevel := writeLogLevel.String()
		assert.Equal(t, expectedLevel, data["level"])
		assert.Equal(t, "test", data["message"])
	}
}
