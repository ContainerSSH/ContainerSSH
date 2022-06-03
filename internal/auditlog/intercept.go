package auditlog

import (
	"io"

    "go.containerssh.io/libcontainerssh/auditlog/message"
)

type interceptingReader struct {
	backend io.Reader
	stream  message.Stream
	channel *loggerChannel
}

func (i *interceptingReader) Read(p []byte) (n int, err error) {
	n, err = i.backend.Read(p)
	if n > 0 {
		i.channel.io(i.stream, p[0:n])
	}
	return n, err
}

type interceptingWriter struct {
	backend io.Writer
	stream  message.Stream
	channel *loggerChannel
}

func (i *interceptingWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		i.channel.io(i.stream, p)
	}
	n, err = i.backend.Write(p)
	return n, err
}

type interceptingReadWriteCloser struct {
	backend io.ReadWriteCloser
	reader interceptingReader
	writer interceptingWriter
}

func (i interceptingReadWriteCloser) Read(p []byte) (int, error ) {
	return i.reader.Read(p)
}

func (i interceptingReadWriteCloser) Write(p []byte) (int, error ) {
	return i.writer.Write(p)
}

func (i interceptingReadWriteCloser) Close() error {
	return i.backend.Close()
}