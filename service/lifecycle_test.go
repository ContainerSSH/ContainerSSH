package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

    "go.containerssh.io/libcontainerssh/service"
)

func TestLifecycle(t *testing.T) {
	l := service.NewLifecycle(newTestService("Test service"))
	starting := make(chan bool)
	running := make(chan bool)
	stopping := make(chan bool)
	stopped := make(chan bool)
	stopExited := make(chan bool)
	l.OnStarting(func(s service.Service, l service.Lifecycle) {
		starting <- true
	})
	l.OnRunning(func(s service.Service, l service.Lifecycle) {
		running <- true
	})
	l.OnStopping(func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
		stopping <- true
	})
	l.OnStopped(func(s service.Service, l service.Lifecycle) {
		stopped <- true
	})
	l.OnCrashed(func(s service.Service, l service.Lifecycle, err error) {
		t.Fail()
	})

	go func() {
		if err := l.Run(); err != nil {
			assert.Fail(t, "service crashed", err)
		}
	}()
	<-starting
	<-running
	go func() {
		l.Stop(context.Background())
		stopExited <- true
	}()
	<-stopping
	<-stopped
	<-stopExited
}

func TestCrash(t *testing.T) {
	s := newTestService("Test service")
	l := service.NewLifecycle(s)
	starting := make(chan bool)
	running := make(chan bool)
	crashed := make(chan bool, 1)
	l.OnStarting(func(s service.Service, l service.Lifecycle) {
		starting <- true
	})
	l.OnRunning(func(s service.Service, l service.Lifecycle) {
		running <- true
	})
	l.OnCrashed(func(s service.Service, l service.Lifecycle, err error) {
		crashed <- true
	})
	l.OnStopped(func(s service.Service, l service.Lifecycle) {
		t.Fail()
	})
	l.OnStopping(func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
		t.Fail()
	})
	go func() {
		if err := l.Run(); err == nil {
			assert.Fail(t, "service did not crash")
		}
	}()
	<-starting
	<-running
	s.Crash()
	<-crashed
}
