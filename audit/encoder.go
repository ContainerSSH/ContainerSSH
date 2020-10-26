package audit

import (
	"github.com/containerssh/containerssh/audit/format/audit"
	"io"
)

type Encoder interface {
	Encode(messages <-chan audit.Message, storage io.Writer)
}
