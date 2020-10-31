package factory

import (
	"fmt"
	"github.com/containerssh/containerssh/audit/format"

	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/format/asciinema"
	"github.com/containerssh/containerssh/audit/none"
	"github.com/containerssh/containerssh/audit/storage/file"
	"github.com/containerssh/containerssh/audit/storage/s3"
	"github.com/containerssh/containerssh/config"
	containersshLog "github.com/containerssh/containerssh/log"
)

func GetStorage(cfg config.AuditConfig, logger containersshLog.Logger) (audit.Storage, error) {
	switch cfg.Storage {
	case config.AuditStorage_None:
		return none.NewStorage(), nil
	case config.AuditStorage_File:
		return file.NewStorage(cfg.File, logger)
	case config.AuditStorage_S3:
		return s3.NewStorage(cfg.S3, logger)
	default:
		return nil, fmt.Errorf("invalid audit storage: %s", cfg.Storage)
	}
}

func GetEncoder(cfg config.AuditConfig, logger containersshLog.Logger) (audit.Encoder, error) {
	switch cfg.Format {
	case config.AuditFormat_None:
		return none.NewEncoder()
	case config.AuditFormat_Asciinema:
		return asciinema.NewEncoder(logger)
	case config.AuditFormat_Audit:
		return format.NewEncoder(logger)
	default:
		return nil, fmt.Errorf("invalid audit log format: %s", cfg.Format)
	}
}

func Get(cfg config.AuditConfig, logger containersshLog.Logger) (audit.Plugin, error) {
	storage, err := GetStorage(cfg, logger)
	if err != nil {
		return nil, err
	}
	encoder, err := GetEncoder(cfg, logger)
	if err != nil {
		return nil, err
	}

	return audit.NewPlugin(logger, storage, encoder), nil
}
