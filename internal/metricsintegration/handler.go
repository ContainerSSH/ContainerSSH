package metricsintegration

import (
	"context"
	"net"
	"sync"

    auth2 "go.containerssh.io/libcontainerssh/auth"
    "go.containerssh.io/libcontainerssh/internal/auth"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/metadata"
)

type metricsHandler struct {
	backend                   sshserver.Handler
	metricsCollector          metrics.Collector
	connectionsMetric         metrics.SimpleGeoCounter
	handshakeSuccessfulMetric metrics.SimpleGeoCounter
	handshakeFailedMetric     metrics.SimpleGeoCounter
	currentConnectionsMetric  metrics.SimpleGeoGauge
}

func (m *metricsHandler) OnReady() error {
	return m.backend.OnReady()
}

func (m *metricsHandler) OnShutdown(shutdownContext context.Context) {
	m.backend.OnShutdown(shutdownContext)
}

func (m *metricsHandler) OnNetworkConnection(
	meta metadata.ConnectionMetadata,
) (sshserver.NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	networkBackend, meta, err := m.backend.OnNetworkConnection(meta)
	if err != nil {
		return networkBackend, meta, err
	}
	m.connectionsMetric.Increment(meta.RemoteAddress.IP)
	m.currentConnectionsMetric.Increment(meta.RemoteAddress.IP)
	return &metricsNetworkHandler{
		client:  net.TCPAddr(meta.RemoteAddress),
		backend: networkBackend,
		handler: m,
		lock:    &sync.Mutex{},
	}, meta, nil
}

type metricsNetworkHandler struct {
	backend      sshserver.NetworkConnectionHandler
	client       net.TCPAddr
	handler      *metricsHandler
	lock         *sync.Mutex
	disconnected bool
}

func (m *metricsNetworkHandler) OnShutdown(shutdownContext context.Context) {
	m.backend.OnShutdown(shutdownContext)
}

func (m *metricsNetworkHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, password []byte) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return m.backend.OnAuthPassword(meta, password)
}

func (m *metricsNetworkHandler) OnAuthPubKey(meta metadata.ConnectionAuthPendingMetadata, pubKey auth2.PublicKey) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return m.backend.OnAuthPubKey(meta, pubKey)
}

func (m *metricsNetworkHandler) OnAuthKeyboardInteractive(
	meta metadata.ConnectionAuthPendingMetadata,
	challenge func(
		instruction string,
		questions sshserver.KeyboardInteractiveQuestions,
	) (answers sshserver.KeyboardInteractiveAnswers, err error),
) (
	response sshserver.AuthResponse,
	metadata metadata.ConnectionAuthenticatedMetadata,
	reason error,
) {
	return m.backend.OnAuthKeyboardInteractive(meta, challenge)
}

func (m *metricsNetworkHandler) OnAuthGSSAPI(meta metadata.ConnectionMetadata) auth.GSSAPIServer {
	return m.backend.OnAuthGSSAPI(meta)
}

func (m *metricsNetworkHandler) OnHandshakeFailed(meta metadata.ConnectionMetadata, reason error) {
	m.handler.handshakeFailedMetric.Increment(m.client.IP)
	m.backend.OnHandshakeFailed(meta, reason)
}

func (m *metricsNetworkHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	metadata metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	connectionHandler, meta, failureReason := m.backend.OnHandshakeSuccess(meta)
	if failureReason != nil {
		m.handler.handshakeFailedMetric.Increment(m.client.IP)
		return connectionHandler, meta, failureReason
	}
	m.handler.handshakeSuccessfulMetric.Increment(m.client.IP)
	return connectionHandler, meta, failureReason
}

func (m *metricsNetworkHandler) OnDisconnect() {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.disconnected {
		m.handler.currentConnectionsMetric.Decrement(m.client.IP)
		m.disconnected = true
	}
	m.backend.OnDisconnect()
}
