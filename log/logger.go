package log

import (
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/message"
)

// Logger The logger interface provides logging facilities on various levels.
type Logger interface {
	// WithLevel returns a copy of the logger for a specified log level. Panics if the log level provided is invalid.
	WithLevel(level config.LogLevel) Logger
	// WithLabel returns a logger with an added label (e.g. username, IP, etc.) Panics if the label name is empty.
	WithLabel(labelName message.LabelName, labelValue message.LabelValue) Logger

	// Debug logs a message at the debug level.
	Debug(message ...interface{})

	// Info logs a message at the info level.
	Info(message ...interface{})

	// Notice logs a message at the notice level.
	Notice(message ...interface{})

	// Warning logs a message at the warning level.
	Warning(message ...interface{})

	// Error logs a message at the error level.
	Error(message ...interface{})

	// Critical logs a message at the critical level.
	Critical(message ...interface{})

	// Alert logs a message at the alert level.
	Alert(message ...interface{})

	// Emergency logs a message at the emergency level.
	Emergency(message ...interface{})

	// Log logs a number of objects or strings to the log.
	Log(v ...interface{})
	// Logf formats a message and logs it.
	Logf(format string, v ...interface{})

	// Rotate triggers the logging backend to close all connections and reopen them to allow for rotating log files.
	Rotate() error
	// Close closes the logging backend.
	Close() error
}

// LoggerFactory is a factory to create a logger on demand
type LoggerFactory interface {
	// Make creates a new logger with the specified configuration and module.
	//
	// - config is the configuration structure.
	//
	// Return:
	//
	// - Logger is the logger created.
	// - error is returned if the configuration was invalid.
	Make(config config.LogConfig) (Logger, error)

	// MustMake is identical to Make but panics if an error happens
	MustMake(config config.LogConfig) Logger
}
