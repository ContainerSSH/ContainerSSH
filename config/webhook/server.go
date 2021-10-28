package webhook

import (
	configuration "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/config"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/service"
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
