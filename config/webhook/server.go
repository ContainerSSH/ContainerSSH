package webhook

import (
	configuration "github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/config"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/service"
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
