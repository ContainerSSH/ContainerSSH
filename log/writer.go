package log

import (
	"io"

	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/message"
)

// Writer is a specialized writer to write a line of log messages.
type Writer interface {
	// Write writes a log message to the output.
	Write(level config.LogLevel, message message.Message) error
	// Rotate attempts to rotate the logs. Has no effect on non-file based loggers.
	Rotate() error

	io.Closer
}
