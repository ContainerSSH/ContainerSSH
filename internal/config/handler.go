package config

import (
    "go.containerssh.io/libcontainerssh/config"
)

// RequestHandler is a generic interface for simplified configuration request handling.
type RequestHandler interface {
	// OnConfig handles configuration requests. It should respond with either an error, resulting in a HTTP 500 response
	// code, or an AppConfig struct, which will be passed back to the client.
	OnConfig(request config.Request) (config.AppConfig, error)
}
