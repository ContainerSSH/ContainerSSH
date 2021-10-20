package backend

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/containerssh/containerssh/config"
	internalConfig "github.com/containerssh/containerssh/internal/config"
	"github.com/containerssh/containerssh/internal/docker"
	"github.com/containerssh/containerssh/internal/kubernetes"
	"github.com/containerssh/containerssh/internal/metrics"
	"github.com/containerssh/containerssh/internal/security"
	"github.com/containerssh/containerssh/internal/sshproxy"
	"github.com/containerssh/containerssh/internal/sshserver"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
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
	remoteAddr net.TCPAddr,
	connectionID string,
) (sshserver.NetworkConnectionHandler, error) {
	return &networkHandler{
		logger: h.logger.
			WithLabel("connectionId", connectionID).
			WithLabel("remoteAddr", remoteAddr.IP.String()),
		rootHandler:  h,
		remoteAddr:   remoteAddr,
		connectionID: connectionID,
		lock:         &sync.Mutex{},
	}, nil
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

func (n *networkHandler) OnAuthPassword(_ string, _ []byte, _ string) (response sshserver.AuthResponse, metadata map[string]string, reason error) {
	return n.authResponse()
}

func (n *networkHandler) authResponse() (sshserver.AuthResponse, map[string]string, error) {
	switch n.rootHandler.authResponse {
	case sshserver.AuthResponseUnavailable:
		return sshserver.AuthResponseUnavailable, nil, fmt.Errorf("the backend handler does not support authentication")
	default:
		return n.rootHandler.authResponse, nil, nil
	}
}

func (n *networkHandler) OnAuthPubKey(_ string, _ string, _ string) (response sshserver.AuthResponse, metadata map[string]string, reason error) {
	return n.authResponse()
}

func (n *networkHandler) OnHandshakeFailed(_ error) {
}

func (n *networkHandler) OnHandshakeSuccess(username string, clientVersion string, metadata map[string]string) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	appConfig, err := n.loadConnectionSpecificConfig(username, metadata)
	if err != nil {
		return nil, err
	}

	backendLogger := n.logger.WithLevel(appConfig.Log.Level).WithLabel("username", username)

	return n.initBackend(username, appConfig, backendLogger, clientVersion, metadata)
}

func (n *networkHandler) initBackend(
	username string,
	appConfig config.AppConfig,
	backendLogger log.Logger,
	version string,
	metadata map[string]string,
) (sshserver.SSHConnectionHandler, error) {
	backend, failureReason := n.getConfiguredBackend(
		appConfig,
		backendLogger,
		n.rootHandler.backendRequestsCounter.WithLabels(metrics.Label(MetricLabelBackend, appConfig.Backend)),
		n.rootHandler.backendErrorCounter.WithLabels(metrics.Label(MetricLabelBackend, appConfig.Backend)),
	)
	if failureReason != nil {
		return nil, failureReason
	}

	// Inject security overlay
	backend, failureReason = security.New(appConfig.Security, backend, n.logger)
	if failureReason != nil {
		return nil, failureReason
	}
	n.backend = backend

	return backend.OnHandshakeSuccess(username, version, metadata)
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
	case "dockerrun":
		//goland:noinspection GoDeprecation
		backend, failureReason = docker.NewDockerRun(
			n.remoteAddr,
			n.connectionID,
			appConfig.DockerRun,
			backendLogger.WithLabel("backend", "dockerrun"),
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
	case "kuberun":
		//goland:noinspection GoDeprecation
		backend, failureReason = kubernetes.NewKubeRun(
			n.remoteAddr,
			n.connectionID,
			appConfig.KubeRun,
			backendLogger.WithLabel("backend", "kuberun"),
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
	username string,
	metadata map[string]string,
) (
	config.AppConfig,
	error,
) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()

	appConfig := config.AppConfig{}
	if err := structutils.Copy(&appConfig, &n.rootHandler.config); err != nil {
		return appConfig, fmt.Errorf("failed to copy application configuration (%w)", err)
	}

	if err := n.rootHandler.configLoader.LoadConnection(
		ctx,
		username,
		n.remoteAddr,
		n.connectionID,
		metadata,
		&appConfig,
	); err != nil {
		return appConfig, fmt.Errorf("failed to load connections-specific configuration (%w)", err)
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
		return appConfig, newErr
	}

	return appConfig, nil
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
