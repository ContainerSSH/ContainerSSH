package none

import (
	"github.com/containerssh/containerssh/audit"
	auditFormat "github.com/containerssh/containerssh/audit/format/audit"
)

func NewPlugin() audit.Plugin {
	return &AuditPlugin{}
}

type AuditPlugin struct {
}

func (a AuditPlugin) Message(_ auditFormat.Message) {
}
