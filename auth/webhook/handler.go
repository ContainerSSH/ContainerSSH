package webhook

import (
	"github.com/containerssh/libcontainerssh/internal/auth"
)

// AuthRequestHandler describes the methods an authentication server has to implement in order to be usable with the
// server component of this package.
type AuthRequestHandler interface {
	auth.Handler
}
