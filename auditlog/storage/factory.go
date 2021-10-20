package storage

import (
	"github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/auditlog"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
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
