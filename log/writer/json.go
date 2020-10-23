package writer

import (
	"encoding/json"
	"github.com/containerssh/containerssh/log"
	"time"
)

type JsonLine struct {
	Time    string          `json:"timestamp"`
	Level   log.LevelString `json:"level"`
	Message string          `json:"message,omitempty"`
	Details interface{}     `json:"details,omitempty"`
}

type JsonLogWriter struct {
}

func NewJsonLogWriter() *JsonLogWriter {
	return &JsonLogWriter{}
}

func (writer *JsonLogWriter) Write(level log.Level, message string) {
	l, err := level.ToString()
	if err != nil {
		panic(err)
	}
	line, err := json.Marshal(JsonLine{
		Time:    time.Now().Format(time.RFC3339),
		Level:   l,
		Message: message,
	})
	if err != nil {
		panic(err)
	}
	println(string(line))
}

func (writer *JsonLogWriter) WriteData(level log.Level, data interface{}) {
	l, err := level.ToString()
	if err != nil {
		panic(err)
	}
	line, err := json.Marshal(JsonLine{
		Time:    time.Now().Format(time.RFC3339),
		Level:   l,
		Details: data,
	})
	if err != nil {
		panic(err)
	}
	println(line)
}
