package kubernetes

import (
	"context"
	"io"
)

// kubernetesExecution is an execution process on either an "exec" process or attached to the main console of a Pod.
type kubernetesExecution interface {
	// resize resizes the current terminal to the given dimensions.
	resize(ctx context.Context, height uint, width uint) error
	// signal sends the given signal to the currently running process. Returns an error if the process is not running,
	// the signal is not known or permitted, or the process ID is not known.
	signal(ctx context.Context, sig string) error
	// run runs the process in question.
	run(
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
		closeWrite func() error,
		onExit func(exitStatus int),
	)
	// done returns a channel that is closed when the program has finished.
	done() <-chan struct{}
	// term notifies the container or execution of an impending termination.
	term(ctx context.Context)
	// kill stops the container or execution.
	kill()
}
