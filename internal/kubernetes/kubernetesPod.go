package kubernetes

import (
	"context"
)

// kubernetesPod is the representation of a created Pod.
type kubernetesPod interface {
	// attach attaches to the Pod on the main console.
	attach(ctx context.Context) (kubernetesExecution, error)

	// createExec creates an execution process for the given program with the given parameters. The passed context is
	// the start context.
	createExec(ctx context.Context, program []string, env map[string]string, tty bool) (kubernetesExecution, error)

	writeFile(ctx context.Context, path string, content []byte) error

	// remove removes the Pod within the given context.
	remove(ctx context.Context) error
}
