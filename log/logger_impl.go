package log

import (
    "go.containerssh.io/libcontainerssh/config"
    messageCodes "go.containerssh.io/libcontainerssh/message"
)

type logger struct {
	level  config.LogLevel
	labels messageCodes.Labels
	writer Writer
	helper func()
}

func (pipeline *logger) Close() error {
	return pipeline.writer.Close()
}

func (pipeline *logger) Rotate() error {
	return pipeline.writer.Rotate()
}

func (pipeline *logger) WithLevel(level config.LogLevel) Logger {
	return &logger{
		level:  level,
		labels: pipeline.labels,
		writer: pipeline.writer,
		helper: pipeline.helper,
	}
}

func (pipeline *logger) WithLabel(labelName messageCodes.LabelName, labelValue messageCodes.LabelValue) Logger {
	newLabels := make(messageCodes.Labels, len(pipeline.labels))
	for k, v := range pipeline.labels {
		newLabels[k] = v
	}
	newLabels[labelName] = labelValue
	return &logger{
		level:  pipeline.level,
		labels: newLabels,
		writer: pipeline.writer,
		helper: pipeline.helper,
	}
}

//region LogFormat

func (pipeline *logger) write(level config.LogLevel, message ...interface{}) {
	pipeline.helper()
	if pipeline.level >= level {
		if len(message) == 0 {
			return
		}
		var msg messageCodes.Message
		if len(message) == 1 {
			switch message[0].(type) {
			case string:
				msg = messageCodes.NewMessage(messageCodes.EUnknownError, message[0].(string))
			default:
				if m, ok := message[0].(messageCodes.Message); ok {
					msg = m
				} else if m, ok := message[0].(error); ok {
					msg = pipeline.wrapError(m)
				} else {
					msg = messageCodes.NewMessage(messageCodes.EUnknownError, "%v", message[0])
				}
			}
		} else {
			msg = messageCodes.NewMessage(messageCodes.EUnknownError, "%v", message)
		}

		for label, value := range pipeline.labels {
			msg = msg.Label(label, value)
		}

		if err := pipeline.writer.Write(level, msg); err != nil {
			panic(err)
		}
	}
}

func (pipeline *logger) writef(level config.LogLevel, format string, args ...interface{}) {
	pipeline.helper()
	if pipeline.level >= level {
		var msg messageCodes.Message

		msg = messageCodes.NewMessage(messageCodes.EUnknownError, format, args...)

		for label, value := range pipeline.labels {
			msg = msg.Label(label, value)
		}

		if err := pipeline.writer.Write(level, msg); err != nil {
			panic(err)
		}
	}
}

func (pipeline *logger) wrapError(err error) messageCodes.Message {
	return messageCodes.Wrap(
		err,
		messageCodes.EUnknownError,
		"An unexpected error has happened.",
	)
}

//endregion

//region Messages

func (pipeline *logger) Emergency(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelEmergency, message...)
}

func (pipeline *logger) Alert(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelAlert, message...)
}

func (pipeline *logger) Critical(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelCritical, message...)
}

func (pipeline *logger) Error(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelError, message...)
}

func (pipeline *logger) Warning(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelWarning, message...)
}

func (pipeline *logger) Notice(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelNotice, message...)
}

func (pipeline *logger) Info(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelInfo, message...)
}

func (pipeline *logger) Debug(message ...interface{}) {
	pipeline.helper()
	pipeline.write(config.LogLevelDebug, message...)
}

//endregion

//region Log

// Log provides a generic log method that logs on the info level.
func (pipeline *logger) Log(args ...interface{}) {
	pipeline.helper()
	pipeline.writef(config.LogLevelInfo, "%v", args...)
}

// Logf provides a generic log method that logs on the info level with formatting.
func (pipeline *logger) Logf(format string, args ...interface{}) {
	pipeline.helper()
	pipeline.writef(config.LogLevelInfo, format, args...)
}

//endregion
