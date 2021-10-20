package webhook

import (
	"github.com/containerssh/containerssh/internal/config"
)

// Client is the interface to fetch a user-specific configuration.
type Client interface {
	config.Client
}
