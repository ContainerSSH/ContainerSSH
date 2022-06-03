package log

import (
	"runtime"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/message"
)

func newGoTest(t *testing.T) Writer {
	return &goTestWriter{
		t: t,
	}
}

type goTestWriter struct {
	t *testing.T
}

func (g *goTestWriter) Write(level config.LogLevel, message message.Message) error {
	g.t.Helper()
	levelString, err := level.Name()
	if err != nil {
		return err
	}

	if testLoggerActive {
		_, file, line, _ := runtime.Caller(3)

		g.t.Logf(
			"\t%s\t%d\t%s\t%s\t%s\n",
			file,
			line,
			levelString,
			message.Code(),
			message.Explanation(),
		)
	} else {
		g.t.Logf("%s\t%s", levelString, message.Explanation())
	}

	return nil
}

func (g *goTestWriter) Rotate() error {
	return nil
}

func (g *goTestWriter) Close() error {
	return nil
}
