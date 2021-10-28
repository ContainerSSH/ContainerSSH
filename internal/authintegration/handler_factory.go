package authintegration

import (
	"fmt"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/auth"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/service"
)

// New creates a new handler that authenticates the users with passwords and public keys.
//goland:noinspection GoUnusedExportedFunction
func New(
	config config.AuthConfig,
	backend sshserver.Handler,
	logger log.Logger,
	metricsCollector metrics.Collector,
	behavior Behavior,
) (sshserver.Handler, service.Service, error) {
	authClient, srv, err := auth.NewClient(config, logger, metricsCollector)
	if err != nil {
		return nil, nil, err
	}
	if backend == nil {
		return nil, nil, fmt.Errorf("the backend parameter to authintegration.New cannot be nil")
	}
	if !behavior.validate() {
		return nil, nil, fmt.Errorf("the behavior field contains an invalid value: %d", behavior)
	}
	return &handler{
		authClient: authClient,
		backend:    backend,
		behavior:   behavior,
	}, srv, nil
}
