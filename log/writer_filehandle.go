package log

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/message"
)

func newFileHandleWriter(fh io.Writer, format config.LogFormat, lock *sync.Mutex) *fileHandleWriter {
	return &fileHandleWriter{
		fh:     fh,
		lock:   lock,
		format: format,
	}
}

type fileHandleWriter struct {
	lock   *sync.Mutex
	fh     io.Writer
	format config.LogFormat
}

func (f *fileHandleWriter) Write(level config.LogLevel, msg message.Message) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	levelString, err := level.Name()
	if err != nil {
		return err
	}
	line, err := f.createLine(levelString, msg)
	if err != nil {
		return message.Wrap(err, message.ELogWriteFailed, "failed to write log message")
	}
	if _, err := f.fh.Write(append(line, '\n')); err != nil {
		return message.Wrap(err, message.ELogWriteFailed, "failed to write log message")
	}
	return nil
}

func (f *fileHandleWriter) Rotate() error {
	return nil
}

func (f *fileHandleWriter) Close() error {
	return nil
}

func (f *fileHandleWriter) createLine(levelString config.LogLevelString, message message.Message) (line []byte, err error) {
	switch f.format {
	case config.LogFormatLJSON:
		line, err = f.createLineLJSON(levelString, message)
		if err != nil {
			return nil, err
		}
	case config.LogFormatText:
		line = f.createLineText(levelString, message)
	default:
		return nil, fmt.Errorf("log format not supported: %s", f.format)
	}
	return line, nil
}

func (f *fileHandleWriter) createLineText(levelString config.LogLevelString, message message.Message) []byte {
	msg := message.Explanation()
	var labels []string
	for labelName, labelValue := range message.Labels() {
		labels = append(labels, fmt.Sprintf("%s=%s", labelName, labelValue))
	}
	if len(labels) > 0 {
		msg += fmt.Sprintf(" (%s)", strings.Join(labels, " "))
	}
	line := []byte(fmt.Sprintf(
		"%s\t%s\t%s\n",
		time.Now().Format(time.RFC3339),
		levelString,
		msg,
	))
	return line
}

func (f *fileHandleWriter) createLineLJSON(levelString config.LogLevelString, message message.Message) (
	[]byte,
	error,
) {
	details := map[string]interface{}{}
	for label, value := range message.Labels() {
		details[string(label)] = value
	}
	line, err := json.Marshal(
		jsonLine{
			Time:    time.Now().Format(time.RFC3339),
			Code:    message.Code(),
			Level:   string(levelString),
			Message: message.Explanation(),
			Details: details,
		},
	)
	if err != nil {
		return nil, err
	}
	return line, nil
}

type jsonLine struct {
	Time    string                 `json:"timestamp"`
	Level   string                 `json:"level"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}
