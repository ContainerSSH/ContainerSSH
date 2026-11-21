package codec

import (
	internalCodec "go.containerssh.io/containerssh/internal/auditlog/codec"
)

// Decoder is a module that is responsible for decoding a binary testdata stream into audit log messages.
type Decoder interface {
	internalCodec.Decoder
}
