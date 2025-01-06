package webhook

import (
	configuration "go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/config"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/service"
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
