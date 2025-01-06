package log

import (
	"io"
)

// NewGoLogWriter creates an adapter for the go logger that writes using the Log method of the logger.
func NewGoLogWriter(backendLogger Logger) io.Writer {
	return &logWriter{
		logger: backendLogger,
	}
}
