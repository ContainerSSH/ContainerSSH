package config

import (
    "go.containerssh.io/libcontainerssh/config"
    http2 "go.containerssh.io/libcontainerssh/http"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
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
