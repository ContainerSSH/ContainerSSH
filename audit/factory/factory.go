package factory

import (
	"fmt"
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/audit/file"
	"github.com/containerssh/containerssh/audit/log"
	"github.com/containerssh/containerssh/audit/none"
	"github.com/containerssh/containerssh/config"
	containersshLog "github.com/containerssh/containerssh/log"
)

func Get(cfg config.AuditConfig, logger containersshLog.Logger) (audit.Plugin, error) {
	switch cfg.Plugin {
	case config.AuditPluginType_None:
		return none.New(), nil
	case config.AuditPluginType_Log:
		return log.New(logger), nil
	case config.AuditPluginType_File:
		return file.New(cfg.File, logger)
	default:
		return nil, fmt.Errorf("invalid audit plugin type: %s", cfg.Plugin)
	}
}
