package auth

import (
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/http"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/message"
)

// NewServer returns a complete HTTP server that responds to the authentication requests.
func NewServer(
	configuration config.HTTPServerConfiguration,
	h Handler,
	logger log.Logger,
) (http.Server, error) {
	return http.NewServer(
		"Auth Server",
		configuration,
		NewHandler(h, logger),
		logger,
		func(url string) {
			logger.Info(message.NewMessage(
				message.MAuthServerAvailable,
				"The authentication server is now available at %s",
				url,
			))
		},
	)
}
