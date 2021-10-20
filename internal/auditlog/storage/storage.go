package storage

import (
	"context"
	"io"
)

// Entry is a storage entry returned from readers
type Entry struct {
	Name     string
	Metadata map[string]string
}

// ReadWriteStorage is a storage that can store as well as retrieve audit logs
type ReadWriteStorage interface {
	ReadableStorage
	WritableStorage
}

// WritableStorage is an audit log storage type that can be written to
type WritableStorage interface {
	// OpenWriter opens a writer for a specific audit log/
	OpenWriter(name string) (Writer, error)
	// Shutdown waits until all uploads are complete and then shuts down the storage. If the shutdownContext expires it
	// will wait for all running audit logs to finish, then abort all uploads and exit immediately.
	Shutdown(shutdownContext context.Context)
}

// ReadableStorage is an audit log storage type that can be read from.
type ReadableStorage interface {
	OpenReader(name string) (io.ReadCloser, error)
	List() (<-chan Entry, <-chan error)
}

// Writer the Writer is a regular WriteCloser with an added function to set the connection metadata for indexing.
type Writer interface {
	io.WriteCloser

	// SetMetadata Set metadata for the audit log. Will be called multiple times, once when user connects and once when the user
	// authenticates.
	//
	// startTime is the time when the connection started in unix timestamp.
	// sourceIp  is the IP address the user connected from.
	// country   is the ISO country code or "XX" if the lookup failed.
	// username  is the username the user entered. The first time this method is called the username will be nil,
	//           may be called subsequently is the user authenticated.
	SetMetadata(startTime int64, sourceIP string, country string, username *string)
}
