package sshserver

import (
	"github.com/containerssh/libcontainerssh/metadata"
)

type conformanceTestHandler struct {
	AbstractHandler

	backend NetworkConnectionHandler
}

func (h *conformanceTestHandler) OnNetworkConnection(meta metadata.ConnectionMetadata) (NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	return h.backend, meta, nil
}
