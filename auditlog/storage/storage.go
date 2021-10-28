package storage

import (
	"github.com/containerssh/libcontainerssh/internal/auditlog/storage"
)

type Storage interface {
	storage.ReadableStorage
}
