package log

import (
	"github.com/containerssh/containerssh/audit/format"
	containersshLog "github.com/containerssh/containerssh/log"
)

type Plugin struct {
	logger containersshLog.Logger
}

func (p *Plugin) Message(msg format.Message) {
	p.logger.DebugF("audit: %v", msg)
}
