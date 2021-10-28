package backend

import (
	"sync"

	"github.com/containerssh/libcontainerssh/config"
	internalConfig "github.com/containerssh/libcontainerssh/internal/config"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
)

// New creates a new backend handler.
//goland:noinspection GoUnusedExportedFunction
func New(
	config config.AppConfig,
	logger log.Logger,
	metricsCollector metrics.Collector,
	defaultAuthResponse sshserver.AuthResponse,
) (sshserver.Handler, error) {
	loader, err := internalConfig.NewHTTPLoader(
		config.ConfigServer,
		logger,
		metricsCollector,
	)
	if err != nil {
		return nil, err
	}

	backendRequestsCounter := metricsCollector.MustCreateCounter(
		MetricNameBackendRequests,
		MetricUnitBackendRequests,
		MetricHelpBackendRequests,
	)
	backendErrorCounter := metricsCollector.MustCreateCounter(
		MetricNameBackendError,
		MetricUnitBackendError,
		MetricHelpBackendError,
	)

	return &handler{
		config:                 config,
		configLoader:           loader,
		authResponse:           defaultAuthResponse,
		metricsCollector:       metricsCollector,
		logger:                 logger,
		backendRequestsCounter: backendRequestsCounter,
		backendErrorCounter:    backendErrorCounter,
		lock:                   &sync.Mutex{},
	}, nil
}
