package storage

import (
    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/auditlog"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
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
