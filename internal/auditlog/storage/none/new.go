package none

import "github.com/containerssh/libcontainerssh/internal/auditlog/storage"

// NewStorage Creates a storage that swallows everything. This can be used for performance.
func NewStorage() storage.WritableStorage {
	return &nopStorage{}
}
