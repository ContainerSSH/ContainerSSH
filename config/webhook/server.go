package webhook

import (
	configuration "go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/internal/config"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/service"
)

// NewServer returns a complete HTTP server that responds to the configuration requests.
//goland:noinspection GoUnusedExportedFunction
func NewServer(
	cfg configuration.HTTPServerConfiguration,
	h ConfigRequestHandler,
	logger log.Logger,
) (service.Service, error) {
	return config.NewServer(
		cfg,
		h,
		logger,
	)
}
