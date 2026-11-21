package codec

import (
	"go.containerssh.io/containerssh/internal/auditlog/codec/binary"
)

// NewBinaryDecoder returns a decoder for the ContainerSSH binary audit log protocol.
func NewBinaryDecoder() Decoder {
	return binary.NewDecoder()
}
