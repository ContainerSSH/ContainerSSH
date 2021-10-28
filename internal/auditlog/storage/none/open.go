package none

import "github.com/containerssh/libcontainerssh/internal/auditlog/storage"

func (s nopStorage) OpenWriter(_ string) (storage.Writer, error) {
	return &nullWriteCloser{}, nil
}
