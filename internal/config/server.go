package config

import (
	"github.com/containerssh/libcontainerssh/config"
	http2 "github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

// NewServer returns a complete HTTP server that responds to the configuration requests.
func NewServer(
	configuration config.HTTPServerConfiguration,
	h RequestHandler,
	logger log.Logger,
) (http2.Server, error) {
	handler, err := NewHandler(h, logger)
	if err != nil {
		return nil, err
	}
	return http2.NewServer(
		"Config Server",
		configuration,
		handler,
		logger,
		func(url string) {
			logger.Info(
				message.NewMessage(
					message.MConfigServerAvailable,
					"The configuration server is now available at %s",
					url,
				))
		},
	)
}
