package service

import (
	"context"
)

// State describes the current state the service is in.
type State string

const (
	// StateStopped means that the service is currently not running.
	StateStopped State = "stopped"
	// StateStarting means that the service is running currently initializing.
	StateStarting State = "starting"
	// StateRunning means that the service is running normally and is serving requests.
	StateRunning State = "running"
	// StateStopping means that the service has received a stop signal and is currently stopping gracefully.
	StateStopping State = "stopping"
	// StateCrashed means that the service has exited with an error.
	StateCrashed State = "crashed"
)

// Lifecycle contains hooks for the Service the Run functions needs to call as it enters each lifecycle stage.
type Lifecycle interface {
	// region Utility

	// Context returns the running context. If the context is canceled the service should stop gracefully.
	Context() context.Context

	// ShouldStop returns true if a shutdown has been triggered.
	ShouldStop() bool

	// ShutdownContext returns the context by which to finish shutting down. When it expires the service should abort
	//                 any currently running interactions. If this method is called outside the "stopping" state it
	//                 will return an empty context.
	ShutdownContext() context.Context

	// State returns the current state of the service.
	State() State

	// Wait waits for the service to enter the "stopped" or "crashed" state. If the service goes into the "crashed"
	//      state it returns the error that caused the crash.
	Wait() error

	// Error returns the error that caused the service to go into the "crashed" state.
	Error() error

	// endregion

	// region Triggers

	// Stop triggers a shutdown of the Service by setting the context to expire. A shutdownContext provides a
	//      deadline for gracefully terminating existing processes.
	Stop(shutdownContext context.Context)

	// Run runs the associated service and returns when complete.
	Run() error

	// endregion

	// region Hook triggers

	// Running must be called by the Service when it is ready to handle user requests.
	Running()

	// Stopping must be called by the Service before stopping to handle user requests. It returns the shutdown context.
	Stopping() context.Context

	// endregion

	// region Hook setup

	// OnStateChange adds a function handler to be called on any state change.
	OnStateChange(func(s Service, l Lifecycle, state State)) Lifecycle

	// OnStarting adds a function handler to be called when the service is about to start.
	OnStarting(func(s Service, l Lifecycle)) Lifecycle

	// OnRunning adds a function handler to be called when the service is ready to serve user requests.
	OnRunning(func(s Service, l Lifecycle)) Lifecycle

	// OnStopping adds a function handler to be called just before the service is starting to shut down. This can be
	//            used to remove the service from a load balancer. Calling this method again adds a second function to
	//            be called. Must be called before Run.
	OnStopping(func(s Service, l Lifecycle, shutdownContext context.Context)) Lifecycle

	// OnStopped adds a function handler to be called after the service has stopped.
	OnStopped(func(s Service, l Lifecycle)) Lifecycle

	// OnCrashed adds a function handler to be called when the service exited with an error.
	OnCrashed(func(s Service, l Lifecycle, err error)) Lifecycle

	// endregion
}
