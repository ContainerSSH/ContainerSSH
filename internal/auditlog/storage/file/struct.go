package file

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
)

type fileStorage struct {
	directory string
	wg        *sync.WaitGroup
}

func (s *fileStorage) Shutdown(_ context.Context) {
	s.wg.Wait()
}

// OpenReader opens a reader for a specific audit log
func (s *fileStorage) OpenReader(name string) (io.ReadCloser, error) {
	// No gosec issue because inclusion is desired here.
	return os.Open(path.Join(s.directory, name)) //nolint:gosec
}

// List lists the available audit logs
func (s *fileStorage) List() (<-chan storage.Entry, <-chan error) {
	result := make(chan storage.Entry)
	errorChannel := make(chan error)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := filepath.Walk(s.directory, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && info.Size() > 0 && !strings.Contains(info.Name(), ".") {
				result <- storage.Entry{
					Name:     info.Name(),
					Metadata: map[string]string{},
				}
			}
			return err
		}); err != nil {
			errorChannel <- err
		}
		close(result)
		close(errorChannel)
	}()
	return result, errorChannel
}

// OpenWriter opens a writer to store an audit log
func (s *fileStorage) OpenWriter(name string) (storage.Writer, error) {
	file, err := os.Create(path.Join(s.directory, name))
	if err != nil {
		return nil, err
	}
	return &writer{
		file: file,
	}, nil
}

type writer struct {
	file *os.File
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

func (w *writer) Close() error {
	return w.file.Close()
}

func (w *writer) SetMetadata(_ int64, _ string, _ string, _ *string) {
}
