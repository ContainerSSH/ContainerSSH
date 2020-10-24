package log

import (
	"github.com/containerssh/containerssh/audit/protocol"
	containersshLog "github.com/containerssh/containerssh/log"
)

type Plugin struct {
	logger containersshLog.Logger
}

func (p *Plugin) Message(msg protocol.Message) {
	p.logger.DebugF("audit: %v", msg)
}
