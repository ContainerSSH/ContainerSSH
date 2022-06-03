package sshserver

import (
    "go.containerssh.io/libcontainerssh/metadata"
)

type conformanceTestHandler struct {
	AbstractHandler

	backend NetworkConnectionHandler
}

func (h *conformanceTestHandler) OnNetworkConnection(meta metadata.ConnectionMetadata) (NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	return h.backend, meta, nil
}
