package log

type Writer interface{
	Write(level Level, message string)
	WriteData(level Level, data interface{})
}
