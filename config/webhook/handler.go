package webhook

import (
    "go.containerssh.io/containerssh/internal/config"
)

// ConfigRequestHandler is a generic interface for simplified configuration request handling.
type ConfigRequestHandler interface {
	config.RequestHandler
}
