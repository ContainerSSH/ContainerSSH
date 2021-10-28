package codec

import (
	internalCodec "github.com/containerssh/libcontainerssh/internal/auditlog/codec"
)

// Decoder is a module that is resonsible for decoding a binary testdata stream into audit log messages.
type Decoder interface {
	internalCodec.Decoder
}
