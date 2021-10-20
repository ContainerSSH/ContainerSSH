package none

import "github.com/containerssh/containerssh/internal/auditlog/storage"

// NewStorage Creates a storage that swallows everything. This can be used for performance.
func NewStorage() storage.WritableStorage {
	return &nopStorage{}
}
