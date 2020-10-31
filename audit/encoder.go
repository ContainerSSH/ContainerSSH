package audit

import (
	"github.com/containerssh/containerssh/audit/format/audit"
)

type Encoder interface {
	Encode(messages <-chan audit.Message, storage StorageWriter)
}
