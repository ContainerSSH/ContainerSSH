package storage

import (
    "go.containerssh.io/containerssh/internal/auditlog/storage"
)

type Storage interface {
	storage.ReadableStorage
}
