package storage

import (
    "go.containerssh.io/libcontainerssh/internal/auditlog/storage"
)

type Storage interface {
	storage.ReadableStorage
}
