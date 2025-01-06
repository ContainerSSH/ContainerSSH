package kubernetes

import (
	"context"
)

// kubernetesClient is a simplified representation of a kubernetes client.
type kubernetesClient interface {
	// createPod creates and starts the configured Pod. May return a Pod even if an error happened.
	// This pod will need to be removed. Passing tty also means that the main console will be prepared for
	// attaching.
	createPod(
		ctx context.Context,
		labels map[string]string,
		annotations map[string]string,
		env map[string]string,
		tty *bool,
		cmd []string,
	) (kubernetesPod, error)
}
