package webhook

import (
	"github.com/containerssh/containerssh/internal/config"
)

// ConfigRequestHandler is a generic interface for simplified configuration request handling.
type ConfigRequestHandler interface {
	config.RequestHandler
}
