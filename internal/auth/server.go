package auth

import (
	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/http"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
)

// NewServer returns a complete HTTP server that responds to the authentication requests.
//goland:noinspection GoUnusedExportedFunction
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
