package none

import "go.containerssh.io/containerssh/internal/auditlog/storage"

func (s nopStorage) OpenWriter(_ string) (storage.Writer, error) {
	return &nullWriteCloser{}, nil
}
