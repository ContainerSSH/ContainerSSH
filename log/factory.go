package log

type LoggerFactory interface{
	Make(config Config) Logger
}
