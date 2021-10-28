package http

import (
	"github.com/containerssh/libcontainerssh/service"
)

// Server is an interface that specifies the minimum requirements for the server.
type Server interface {
	service.Service
}
