package config

import (
	"github.com/containerssh/containerssh/config"
	http2 "github.com/containerssh/containerssh/http"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/message"
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
