package sshserver

import (
	"fmt"
	"sync"

	config2 "github.com/containerssh/containerssh/config"
	"github.com/containerssh/containerssh/internal/structutils"
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/service"
)

var testServerLock = &sync.Mutex{}
var nextPort = 2222

// NewTestServer is a simplified API to start and stop a test server. The test server always listens on 127.0.0.1:2222
func NewTestServer(handler Handler, logger log.Logger) TestServer {
	config := config2.SSHConfig{}
	structutils.Defaults(&config)

	testServerLock.Lock()
	config.Listen = fmt.Sprintf("127.0.0.1:%d", nextPort)
	nextPort++
	testServerLock.Unlock()
	if err := config.GenerateHostKey(); err != nil {
		panic(err)
	}
	svc, err := New(config, handler, logger)
	if err != nil {
		panic(err)
	}
	lifecycle := service.NewLifecycle(svc)
	started := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			started <- struct{}{}
		})

	return &testServerImpl{
		config:    config,
		lifecycle: lifecycle,
		started:   started,
	}
}
