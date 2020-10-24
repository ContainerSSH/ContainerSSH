package none

import (
	"github.com/containerssh/containerssh/audit/format"
)

type Plugin struct {
}

func (p *Plugin) Message(_ format.Message) {
}
