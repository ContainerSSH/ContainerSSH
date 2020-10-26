package audit

import (
	"github.com/containerssh/containerssh/audit/format/audit"
	"io"
)

type interceptingReader struct {
	backend io.Reader
	stream  audit.Stream
	channel *Channel
}

func (i *interceptingReader) Read(p []byte) (n int, err error) {
	n, err = i.backend.Read(p)
	i.channel.Message(audit.MessageType_IO, audit.MessageIO{
		Stream: i.stream,
		Data:   p[0:n],
	})
	return n, err
}

type interceptingWriter struct {
	backend io.Writer
	stream  audit.Stream
	channel *Channel
}

func (i *interceptingWriter) Write(p []byte) (n int, err error) {
	i.channel.Message(audit.MessageType_IO, audit.MessageIO{
		Stream: i.stream,
		Data:   p,
	})
	n, err = i.backend.Write(p)
	return n, err
}
