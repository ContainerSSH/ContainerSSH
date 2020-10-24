package none

import (
	"github.com/containerssh/containerssh/audit/protocol"
)

type Plugin struct {
}

func (p *Plugin) Message(_ protocol.Message) {
}
