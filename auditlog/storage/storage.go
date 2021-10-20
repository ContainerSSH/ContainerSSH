package storage

import (
	"github.com/containerssh/containerssh/internal/auditlog/storage"
)

type Storage interface {
	storage.ReadableStorage
}
