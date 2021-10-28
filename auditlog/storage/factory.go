package storage

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auditlog"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

func New(cfg config.AuditLogConfig, logger log.Logger) (Storage, error) {
	storage, err := auditlog.NewStorage(cfg, logger)
	if err != nil {
		return nil, err
	}
	readableStorage, ok := storage.(Storage)
	if !ok {
		return nil, message.NewMessage(
			message.EAuditLogStorageNotReadable,
			"The specified storage (%s) is not readable.",
			cfg.Storage,
		)
	}
	return readableStorage, nil
}
