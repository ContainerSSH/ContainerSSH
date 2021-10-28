package containerssh

import (
	"github.com/containerssh/libcontainerssh/service"
)

// Service is the core ContainerSSH service.
type Service interface {
	service.Service

	// RotateLogs closes the currently open logs and reopens them to allow for log rotation.
	RotateLogs() error
}
