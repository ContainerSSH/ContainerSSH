package log

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/message"
)

func newSyslogWriter(config config.SyslogConfig, format config.LogFormat) (Writer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &syslogWriter{
		lock:       &sync.Mutex{},
		connection: config.Connection,
		config:     config,
		format:     format,
	}, nil
}

type syslogWriter struct {
	connection net.Conn
	config     config.SyslogConfig
	lock       *sync.Mutex
	format     config.LogFormat
}

func (s *syslogWriter) Write(level config.LogLevel, messageText message.Message) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	facilityNumber, err := s.config.Facility.Number()
	if err != nil {
		return err
	}
	pri := int64(facilityNumber)*8 + int64(level)
	t := time.Now()
	timestamp := fmt.Sprintf(
		"%s %2d %02d:%02d:%02d",
		t.Format("Feb"),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second())
	tag := s.config.Tag
	if tag == "" {
		tag = "ContainerSSH"
	}
	if s.config.Pid {
		tag += fmt.Sprintf("[%d]", os.Getpid())
	}
	msg, err := s.createMessage(messageText)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("<%d>%s %s: %s\n", pri, timestamp, tag, msg)
	if _, err = s.connection.Write([]byte(line)); err != nil {
		return message.Wrap(err, message.ELogWriteFailed, "failed to write to syslog socket")
	}
	return nil
}

func (s *syslogWriter) createMessage(message message.Message) (line []byte, err error) {
	switch s.format {
	case config.LogFormatLJSON:
		details := map[string]interface{}{}
		for label, value := range message.Labels() {
			details[string(label)] = value
		}
		line, err = json.Marshal(
			syslogJsonLine{
				Code:    message.Code(),
				Message: message.Explanation(),
				Details: details,
			},
		)
		if err != nil {
			return nil, err
		}
	case config.LogFormatText:
		msg := message.Explanation()
		var labels []string
		for labelName, labelValue := range message.Labels() {
			labels = append(labels, fmt.Sprintf("%s=%s", labelName, labelValue))
		}
		if len(labels) > 0 {
			msg += fmt.Sprintf(" (%s)", strings.Join(labels, " "))
		}
		line = []byte(msg)
	default:
		return nil, fmt.Errorf("log format not supported: %s", s.format)
	}
	return line, nil
}

type syslogJsonLine struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

func (s *syslogWriter) Rotate() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.config.Validate(); err != nil {
		return err
	}
	if err := s.connection.Close(); err != nil {
		return message.Wrap(err, message.ELogRotateFailed, "failed to close old syslog connection")
	}
	s.connection = s.config.Connection
	return nil
}

func (s *syslogWriter) Close() error {
	return s.connection.Close()
}
