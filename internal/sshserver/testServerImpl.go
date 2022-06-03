package sshserver

import (
	"context"
	"fmt"
	"time"

    config2 "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/service"
)

type testServerImpl struct {
	config    config2.SSHConfig
	lifecycle service.Lifecycle
	started   chan struct{}
}

func (t *testServerImpl) GetListen() string {
	return t.config.Listen
}

func (t *testServerImpl) GetHostKey() string {
	return t.config.HostKeys[0]
}

func (t *testServerImpl) Start() {
	if t.lifecycle.State() != service.StateStopped {
		panic(fmt.Errorf("server already running"))
	}
	go func() {
		_ = t.lifecycle.Run()
	}()
	<-t.started
}

func (t *testServerImpl) Stop(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	t.lifecycle.Stop(ctx)
}
