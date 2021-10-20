package service

import (
	"context"
	"sync"
)

// NewLifecycle creates a new lifecycle for the specified service. The lifecycle is responsible for managing the start
// and stop of the service.
func NewLifecycle(service Service) Lifecycle {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &lifecycle{
		service:         service,
		state:           StateStopped,
		mutex:           &sync.Mutex{},
		runningContext:  ctx,
		cancelRun:       cancelFunc,
		shutdownContext: context.Background(),
	}
}

// NewLifecycleFactory creates a new default factory for lifecycles.
func NewLifecycleFactory() LifecycleFactory {
	return &lifecycleFactory{}
}

// LifecycleFactory is an interface to create lifecycle objects in pools.
type LifecycleFactory interface {
	Make(service Service) Lifecycle
}

type lifecycleFactory struct {
}

func (l *lifecycleFactory) Make(service Service) Lifecycle {
	return NewLifecycle(service)
}
