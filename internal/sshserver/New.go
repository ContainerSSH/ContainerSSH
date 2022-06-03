package sshserver

import (
	"sync"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/log"
)

// New creates a new SSH server ready to be run. It may return an error if the configuration is invalid.
func New(cfg config.SSHConfig, handler Handler, logger log.Logger) (Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	hostKeys, err := cfg.LoadHostKeys()
	if err != nil {
		return nil, err
	}
	return &serverImpl{
		cfg:          cfg,
		handler:      handler,
		logger:       logger,
		wg:           &sync.WaitGroup{},
		lock:         &sync.Mutex{},
		listenSocket: nil,
		hostKeys:     hostKeys,
		shutdownHandlers: &shutdownRegistry{
			lock:      &sync.Mutex{},
			callbacks: map[string]shutdownHandler{},
		},
	}, nil
}
