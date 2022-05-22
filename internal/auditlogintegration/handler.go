package auditlogintegration

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/containerssh/libcontainerssh/auditlog/message"
	"github.com/containerssh/libcontainerssh/internal/auditlog"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/metadata"
)

type handler struct {
	logger  auditlog.Logger
	backend sshserver.Handler
}

func (h *handler) OnReady() error {
	return h.backend.OnReady()
}

func (h *handler) OnShutdown(shutdownContext context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		h.backend.OnShutdown(shutdownContext)
	}()
	go func() {
		defer wg.Done()
		h.logger.Shutdown(shutdownContext)
	}()
	wg.Wait()
}

func (h *handler) OnNetworkConnection(meta metadata.ConnectionMetadata) (
	sshserver.NetworkConnectionHandler,
	metadata.ConnectionMetadata,
	error,
) {
	backend, meta, err := h.backend.OnNetworkConnection(meta)
	if err != nil {
		return nil, meta, err
	}
	auditConnection, err := h.logger.OnConnect(message.ConnectionID(meta.ConnectionID), net.TCPAddr(meta.RemoteAddress))
	if err != nil {
		return nil, meta, fmt.Errorf(
			"failed to initialize audit logger for connection from %s (%w)",
			meta.RemoteAddress.IP.String(),
			err,
		)
	}

	return &networkConnectionHandler{
		backend: backend,
		audit:   auditConnection,
	}, meta, nil
}
