package backend

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	auth2 "github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/config"
	internalConfig "github.com/containerssh/libcontainerssh/internal/config"
	"github.com/containerssh/libcontainerssh/internal/docker"
	"github.com/containerssh/libcontainerssh/internal/kubernetes"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/security"
	"github.com/containerssh/libcontainerssh/internal/sshproxy"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/containerssh/libcontainerssh/metadata"
)

type handler struct {
	sshserver.AbstractHandler

	config                 config.AppConfig
	configLoader           internalConfig.Loader
	authResponse           sshserver.AuthResponse
	metricsCollector       metrics.Collector
	logger                 log.Logger
	backendRequestsCounter metrics.Counter
	backendErrorCounter    metrics.Counter
	lock                   *sync.Mutex
}

func (h *handler) OnNetworkConnection(
	meta metadata.ConnectionMetadata,
) (sshserver.NetworkConnectionHandler, metadata.ConnectionMetadata, error) {
	return &networkHandler{
		logger: h.logger.
			WithLabel("connectionId", meta.ConnectionID).
			WithLabel("remoteAddr", meta.RemoteAddress.IP.String()),
		rootHandler:  h,
		remoteAddr:   net.TCPAddr(meta.RemoteAddress),
		connectionID: meta.ConnectionID,
		lock:         &sync.Mutex{},
	}, meta, nil
}

type networkHandler struct {
	sshserver.AbstractNetworkConnectionHandler

	rootHandler  *handler
	remoteAddr   net.TCPAddr
	connectionID string
	backend      sshserver.NetworkConnectionHandler
	lock         *sync.Mutex
	logger       log.Logger
}

func (n *networkHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, _ []byte) (
	sshserver.AuthResponse,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return n.authResponse(meta)
}

func (n *networkHandler) authResponse(meta metadata.ConnectionAuthPendingMetadata) (
	sshserver.AuthResponse,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	switch n.rootHandler.authResponse {
	case sshserver.AuthResponseUnavailable:
		return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf("the backend handler does not support authentication")
	case sshserver.AuthResponseSuccess:
		return sshserver.AuthResponseSuccess, meta.Authenticated(meta.Username), nil
	default:
		return sshserver.AuthResponseFailure, meta.AuthFailed(), nil
	}
}

func (n *networkHandler) OnAuthPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	_ auth2.PublicKey,
) (response sshserver.AuthResponse, metadata metadata.ConnectionAuthenticatedMetadata, reason error) {
	return n.authResponse(meta)
}

func (n *networkHandler) OnHandshakeFailed(metadata.ConnectionMetadata, error) {
}

func (n *networkHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	resultMeta metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	appConfig, newMeta, err := n.loadConnectionSpecificConfig(meta)
	if err != nil {
		return nil, meta, err
	}

	backendLogger := n.logger.WithLevel(appConfig.Log.Level).WithLabel(
		"username",
		meta.Username,
	).WithLabel("authenticatedUsername", meta.AuthenticatedUsername)

	return n.initBackend(newMeta, appConfig, backendLogger)
}

func (n *networkHandler) initBackend(
	meta metadata.ConnectionAuthenticatedMetadata,
	appConfig config.AppConfig,
	backendLogger log.Logger,
) (sshserver.SSHConnectionHandler, metadata.ConnectionAuthenticatedMetadata, error) {
	backend, failureReason := n.getConfiguredBackend(
		appConfig,
		backendLogger,
		n.rootHandler.backendRequestsCounter.WithLabels(metrics.Label(MetricLabelBackend, string(appConfig.Backend))),
		n.rootHandler.backendErrorCounter.WithLabels(metrics.Label(MetricLabelBackend, string(appConfig.Backend))),
	)
	if failureReason != nil {
		return nil, meta, failureReason
	}

	// Inject security overlay
	backend, failureReason = security.New(appConfig.Security, backend, n.logger)
	if failureReason != nil {
		return nil, meta, failureReason
	}
	n.backend = backend

	return backend.OnHandshakeSuccess(meta)
}

func (n *networkHandler) getConfiguredBackend(
	appConfig config.AppConfig,
	backendLogger log.Logger,
	backendRequestsCounter metrics.Counter,
	backendErrorCounter metrics.Counter,
) (backend sshserver.NetworkConnectionHandler, failureReason error) {
	switch appConfig.Backend {
	case "docker":
		backend, failureReason = docker.New(
			n.remoteAddr,
			n.connectionID,
			appConfig.Docker,
			backendLogger.WithLabel("backend", "docker"),
			backendRequestsCounter,
			backendErrorCounter,
		)
	case "kubernetes":
		backend, failureReason = kubernetes.New(
			n.remoteAddr,
			n.connectionID,
			appConfig.Kubernetes,
			backendLogger.WithLabel("backend", "kubernetes"),
			backendRequestsCounter,
			backendErrorCounter,
		)
	case "sshproxy":
		backend, failureReason = sshproxy.New(
			n.remoteAddr,
			n.connectionID,
			appConfig.SSHProxy,
			backendLogger.WithLabel("backend", "sshproxy"),
			backendRequestsCounter,
			backendErrorCounter,
		)
	default:
		failureReason = fmt.Errorf("invalid backend: %s", appConfig.Backend)
	}
	return backend, failureReason
}

func (n *networkHandler) loadConnectionSpecificConfig(
	meta metadata.ConnectionAuthenticatedMetadata,
) (
	config.AppConfig,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()

	appConfig := config.AppConfig{}
	if err := structutils.Copy(&appConfig, &n.rootHandler.config); err != nil {
		return appConfig, meta, fmt.Errorf("failed to copy application configuration (%w)", err)
	}

	newMeta, err := n.rootHandler.configLoader.LoadConnection(
		ctx,
		meta,
		&appConfig,
	)
	if err != nil {
		return appConfig, meta, fmt.Errorf("failed to load connections-specific configuration (%w)", err)
	}

	if err := appConfig.Validate(true); err != nil {
		newErr := fmt.Errorf("configuration server returned invalid configuration (%w)", err)
		n.rootHandler.logger.Error(
			message.Wrap(
				err,
				message.EBackendConfig,
				"configuration server returned invalid configuration",
			),
		)
		return appConfig, meta, newErr
	}

	return appConfig, meta.Merge(newMeta), nil
}

func (n *networkHandler) OnDisconnect() {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.backend != nil {
		n.backend.OnDisconnect()
		n.backend = nil
	}
}

func (n *networkHandler) OnShutdown(shutdownContext context.Context) {
	n.lock.Lock()
	if n.backend != nil {
		backend := n.backend
		n.lock.Unlock()
		backend.OnShutdown(shutdownContext)
	} else {
		n.lock.Unlock()
	}
}
