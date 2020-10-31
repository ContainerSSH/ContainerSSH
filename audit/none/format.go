package none

import (
	"github.com/containerssh/containerssh/audit"
	auditFormat "github.com/containerssh/containerssh/audit/format/audit"
)

func NewEncoder() (audit.Encoder, error) {
	return &Encoder{}, nil
}

type Encoder struct {
}

func (e Encoder) Encode(messages <-chan auditFormat.Message, _ audit.StorageWriter) {
	for {
		msg, ok := <-messages
		if !ok {
			break
		}
		if msg.MessageType == auditFormat.MessageType_Disconnect {
			break
		}
	}
}
