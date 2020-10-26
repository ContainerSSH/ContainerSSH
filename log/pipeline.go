package log

import "fmt"

type LoggerPipeline struct {
	config Config
	writer Writer
}

type LoggerPipelineFactory struct {
	writer Writer
}

func NewLoggerPipelineFactory(writer Writer) LoggerFactory {
	return &LoggerPipelineFactory{
		writer: writer,
	}
}

func (f *LoggerPipelineFactory) Make(config Config) Logger {
	return NewLoggerPipeline(
		config,
		f.writer,
	)
}

func NewLoggerPipeline(config Config, writer Writer) *LoggerPipeline {
	return &LoggerPipeline{config: config, writer: writer}
}

// region Write
func (pipeline *LoggerPipeline) write(level Level, message string) {
	if pipeline.config.level >= level {
		pipeline.writer.Write(level, message)
	}
}

func (pipeline *LoggerPipeline) writeE(level Level, err error) {
	if pipeline.config.level >= level {
		pipeline.writer.WriteData(level, err.Error())
	}
}

func (pipeline *LoggerPipeline) writeF(level Level, format string, args ...interface{}) {
	if pipeline.config.level >= level {
		pipeline.writer.Write(level, fmt.Sprintf(format, args...))
	}
}

func (pipeline *LoggerPipeline) writeD(level Level, data interface{}) {
	if pipeline.config.level >= level {
		pipeline.writer.WriteData(level, data)
	}
}

// endregion

// region Emergency
func (pipeline *LoggerPipeline) Emergency(message string) {
	pipeline.write(LevelEmergency, message)
}

func (pipeline *LoggerPipeline) EmergencyE(err error) {
	pipeline.writeE(LevelEmergency, err)
}

func (pipeline *LoggerPipeline) EmergencyD(data interface{}) {
	pipeline.writeD(LevelEmergency, data)
}

func (pipeline *LoggerPipeline) EmergencyF(format string, args ...interface{}) {
	pipeline.writeF(LevelEmergency, format, args...)
}

// endregion

// region Alert
func (pipeline *LoggerPipeline) Alert(message string) {
	pipeline.write(LevelAlert, message)
}

func (pipeline *LoggerPipeline) AlertE(err error) {
	pipeline.writeE(LevelAlert, err)
}

func (pipeline *LoggerPipeline) AlertD(data interface{}) {
	pipeline.writeD(LevelAlert, data)
}

func (pipeline *LoggerPipeline) AlertF(format string, args ...interface{}) {
	pipeline.writeF(LevelAlert, format, args...)
}

// endregion

// region Critical
func (pipeline *LoggerPipeline) Critical(message string) {
	pipeline.write(LevelCritical, message)
}

func (pipeline *LoggerPipeline) CriticalE(err error) {
	pipeline.writeE(LevelCritical, err)
}

func (pipeline *LoggerPipeline) CriticalD(data interface{}) {
	pipeline.writeD(LevelCritical, data)
}

func (pipeline *LoggerPipeline) CriticalF(format string, args ...interface{}) {
	pipeline.writeF(LevelCritical, format, args...)
}

// endregion

// region Error
func (pipeline *LoggerPipeline) Error(message string) {
	pipeline.write(LevelError, message)
}

func (pipeline *LoggerPipeline) ErrorE(err error) {
	pipeline.writeE(LevelError, err)
}

func (pipeline *LoggerPipeline) ErrorD(data interface{}) {
	pipeline.writeD(LevelError, data)
}

func (pipeline *LoggerPipeline) ErrorF(format string, args ...interface{}) {
	pipeline.writeF(LevelError, format, args...)
}

// endregion

// region Warning
func (pipeline *LoggerPipeline) Warning(message string) {
	pipeline.write(LevelWarning, message)
}

func (pipeline *LoggerPipeline) WarningE(err error) {
	pipeline.writeE(LevelWarning, err)
}

func (pipeline *LoggerPipeline) WarningD(data interface{}) {
	pipeline.writeD(LevelWarning, data)
}

func (pipeline *LoggerPipeline) WarningF(format string, args ...interface{}) {
	pipeline.writeF(LevelWarning, format, args...)
}

// endregion

// region Notice
func (pipeline *LoggerPipeline) Notice(message string) {
	pipeline.write(LevelNotice, message)
}

func (pipeline *LoggerPipeline) NoticeE(err error) {
	pipeline.writeE(LevelNotice, err)
}

func (pipeline *LoggerPipeline) NoticeD(data interface{}) {
	pipeline.writeD(LevelNotice, data)
}

func (pipeline *LoggerPipeline) NoticeF(format string, args ...interface{}) {
	pipeline.writeF(LevelNotice, format, args...)
}

// endregion

// region Info
func (pipeline *LoggerPipeline) Info(message string) {
	pipeline.write(LevelInfo, message)
}

func (pipeline *LoggerPipeline) InfoE(err error) {
	pipeline.writeE(LevelInfo, err)
}

func (pipeline *LoggerPipeline) InfoD(data interface{}) {
	pipeline.writeD(LevelInfo, data)
}

func (pipeline *LoggerPipeline) InfoF(format string, args ...interface{}) {
	pipeline.writeF(LevelInfo, format, args...)
}

// endregion

// region Debug
func (pipeline *LoggerPipeline) Debug(message string) {
	pipeline.write(LevelDebug, message)
}

func (pipeline *LoggerPipeline) DebugE(err error) {
	pipeline.writeE(LevelDebug, err)
}

func (pipeline *LoggerPipeline) DebugD(data interface{}) {
	pipeline.writeD(LevelDebug, data)
}

func (pipeline *LoggerPipeline) DebugF(format string, args ...interface{}) {
	pipeline.writeF(LevelDebug, format, args...)
}

// endregion

// region Log
func (pipeline *LoggerPipeline) Log(args ...interface{}) {
	pipeline.writeF(LevelInfo, "%v", args)
}

//endregion
