package containerssh_test

import (
	"context"

	"github.com/containerssh/service"
)

type SimpleLifecycle struct {
	lifecycle service.Lifecycle
	running chan struct{}
	stopped chan struct{}
}

func NewSimpleLifecycle(srv service.Service) *SimpleLifecycle {
	l := &SimpleLifecycle{
		lifecycle: service.NewLifecycle(srv),
		running: make(chan struct{}, 1),
		stopped: make(chan struct{}, 1),
	}
	l.lifecycle.OnRunning(
		func(service.Service, service.Lifecycle) {
			l.running <- struct{}{}
		}).OnStopped(
		func(service.Service, service.Lifecycle) {
			l.stopped <- struct{}{}
		})
	return l
}

func (s *SimpleLifecycle) Start() error {
	go func() {
		_ = s.lifecycle.Run()
	}()
	select {
	case <-s.running:
		return nil
	case <-s.stopped:
		return s.lifecycle.Wait()
	}
}

func (s *SimpleLifecycle) Stop() error {
	s.lifecycle.Stop(context.Background())
	<-s.stopped
	return s.lifecycle.Wait()
}
