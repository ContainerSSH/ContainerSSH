package sshserver

import (
	"net"
)

type conformanceTestHandler struct {
	AbstractHandler

	backend NetworkConnectionHandler
}

func (h *conformanceTestHandler) OnNetworkConnection(_ net.TCPAddr, _ string) (NetworkConnectionHandler, error) {
	return h.backend, nil
}
