package none

import (
	"github.com/containerssh/containerssh/audit"
)

func NewStorage() audit.Storage {
	return &Storage{}
}

type Storage struct {
}

func (s Storage) Open(_ string) (audit.StorageWriter, error) {
	return &NullWriteCloser{}, nil
}

type NullWriteCloser struct {
}

func (w *NullWriteCloser) SetMetadata(_ int64, _ string, _ *string) {
}

func (w *NullWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *NullWriteCloser) Close() error {
	return nil
}
