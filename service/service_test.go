package service_test

import (
	"errors"

    "go.containerssh.io/libcontainerssh/service"
)

type testService struct {
	crash        chan bool
	name         string
	crashStartup bool
}

func (t *testService) String() string {
	return t.name
}

func (t *testService) RunWithLifecycle(lifecycle service.Lifecycle) error {
	if t.crashStartup {
		return errors.New("crash")
	}
	lifecycle.Running()
	ctx := lifecycle.Context()
	select {
	case <-ctx.Done():
		lifecycle.Stopping()
		return nil
	case <-t.crash:
		return errors.New("crash")
	}
}

func (t *testService) CrashStartup() {
	t.crashStartup = true
}

func (t *testService) Crash() {
	select {
	case t.crash <- true:
	default:
	}
}

func newTestService(name string) *testService {
	return &testService{
		name:  name,
		crash: make(chan bool, 1),
	}
}
