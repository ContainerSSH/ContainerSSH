package log

import (
	"bytes"
	"strings"
)

type logWriter struct {
	logger Logger
	buf    bytes.Buffer
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if bytes.Equal([]byte{b}, []byte("\n")) {
			l.logger.Log(strings.TrimSpace(l.buf.String()))
			l.buf.Reset()
		} else {
			l.buf.Write([]byte{b})
		}
	}
	if l.buf.Len() > 0 {
		l.logger.Log(strings.TrimSpace(l.buf.String()))
	}
	return len(p), nil
}
