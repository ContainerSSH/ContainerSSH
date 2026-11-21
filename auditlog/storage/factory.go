package storage

import (
    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/auditlog"
    "go.containerssh.io/containerssh/log"
    "go.containerssh.io/containerssh/message"
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
