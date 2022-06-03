package codec

import (
	"io"

    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
)

// NewStorageWriterProxy Creates a storage writer that proxies to a traditional writer and drops the metadata.
func NewStorageWriterProxy(backend io.WriteCloser) storage.Writer {
	return &storageWriterProxy{backend: backend}
}

type storageWriterProxy struct {
	backend io.WriteCloser
}

func (s *storageWriterProxy) Write(p []byte) (n int, err error) {
	return s.backend.Write(p)
}

func (s *storageWriterProxy) Close() error {
	return s.backend.Close()
}

func (s *storageWriterProxy) SetMetadata(_ int64, _ string, _ string, _ *string) {
	// No metadata storage
}
