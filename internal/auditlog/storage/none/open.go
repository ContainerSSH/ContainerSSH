package none

import "github.com/containerssh/containerssh/internal/auditlog/storage"

func (s nopStorage) OpenWriter(_ string) (storage.Writer, error) {
	return &nullWriteCloser{}, nil
}
