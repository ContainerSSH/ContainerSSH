// Package log provides a Logger interface that can log all kinds of messages and errors, including the Message
// interface from the message package. It provides the following methods for logging:
//
//     logger.Debug(message ...interface{})
//     logger.Info(message ...interface{})
//     logger.Notice(message ...interface{})
//     logger.Warning(message ...interface{})
//     logger.Error(message ...interface{})
//     logger.Critical(message ...interface{})
//     logger.Alert(message ...interface{})
//     logger.Emergency(message ...interface{})
//
// You are encouraged to pass a message.Message implementation to it.
// We also provide the following compatibility methods which log at the info level.
//
//     logger.Log(v ...interface{})
//     logger.Logf(format string, v ...interface{})
//
// We provide a method to create a child logger that has a different minimum log level. Messages below this level will
// be discarded:
//
//     newLogger := logger.WithLevel(log.LevelInfo)
//
// We can also create a new logger copy with default labels added:
//
//     newLogger := logger.WithLabel("label name", "label value")
//
// Finally, the logger also supports calling the Rotate() and Close() methods. Rotate() instructs the output to
// close all handles and reopen them to facilitate rotating logs. Close() permanently closes the writer.
//
// Creating a logger
//
// The Logger interface is intended for generic implementations. The default implementation can be created as follows:
//
//     logger, err := log.NewLogger(config)
//
// Alternatively, you can also use the log.MustNewLogger method to skip having to deal with the error. (It will
// panic if an error happens.)
//
// If you need a factory you can use the log.LoggerFactory interface and the log.NewLoggerFactory to create a factory
// you can pass around. The Make(config) method will make a new logger when needed.
package log
