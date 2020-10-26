package none

import (
	"github.com/containerssh/containerssh/audit"
	"io"
)

func NewStorage() audit.Storage {
	return &Storage{}
}

type Storage struct {
}

func (s Storage) Open(_ string) (io.WriteCloser, error) {
	return &NullWriteCloser{}, nil
}

type NullWriteCloser struct {
}

func (n2 NullWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (n2 NullWriteCloser) Close() error {
	return nil
}
