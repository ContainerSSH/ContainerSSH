package webhook

import (
	"github.com/containerssh/libcontainerssh/internal/config"
)

// ConfigRequestHandler is a generic interface for simplified configuration request handling.
type ConfigRequestHandler interface {
	config.RequestHandler
}
