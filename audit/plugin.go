package audit

import (
	"github.com/containerssh/containerssh/audit/format/audit"
)

// The audit plugin has the ability to log all events happening in the container SSH.
type Plugin interface {
	Message(message audit.Message)
}
