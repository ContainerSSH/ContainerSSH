package audit

import (
	"io"
)

// The StorageWriter is a regular WriteCloser with an added function to set the connection metadata for indexing.
type StorageWriter interface {
	io.WriteCloser

	// Set metadata for the audit log. Can be called multiple times.
	//
	// startTime is the time when the connection started in unix timestamp
	// sourceIp  is the IP address the user connected from
	// username  is the username the user entered. The first time this method is called the username will be nil,
	//           may be called subsequently is the user authenticated.
	SetMetadata(startTime int64, sourceIp string, username *string)
}

type Storage interface {
	Open(name string) (StorageWriter, error)
}

func NewStorageWriterProxy(backend io.WriteCloser, err error) (StorageWriter, error) {
	if err != nil {
		return nil, err
	}
	return &StorageWriterProxy{
		backend: backend,
	}, nil
}

type StorageWriterProxy struct {
	backend io.WriteCloser
}

func (s *StorageWriterProxy) Write(p []byte) (n int, err error) {
	return s.backend.Write(p)
}

func (s *StorageWriterProxy) Close() error {
	return s.backend.Close()
}

func (s *StorageWriterProxy) SetMetadata(_ int64, _ string, _ *string) {
	// No metadata storage
}
