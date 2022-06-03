package http

import (
    "go.containerssh.io/libcontainerssh/service"
)

// Server is an interface that specifies the minimum requirements for the server.
type Server interface {
	service.Service
}
