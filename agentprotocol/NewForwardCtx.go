package agentprotocol

import (
	"io"

    log "go.containerssh.io/containerssh/log"
)

func NewForwardCtx(fromBackend io.Reader, toBackend io.Writer, logger log.Logger) *ForwardCtx {
	return &ForwardCtx{
		fromBackend: fromBackend,
		toBackend: toBackend,
		logger: logger,
	}
}