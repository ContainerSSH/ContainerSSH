package none

import "go.containerssh.io/libcontainerssh/internal/auditlog/storage"

func (s nopStorage) OpenWriter(_ string) (storage.Writer, error) {
	return &nullWriteCloser{}, nil
}
