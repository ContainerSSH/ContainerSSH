package webhook

import (
    "go.containerssh.io/libcontainerssh/internal/auth"
)

// AuthRequestHandler describes the methods an authentication server has to implement in order to be usable with the
// server component of this package.
type AuthRequestHandler interface {
	auth.Handler
}
