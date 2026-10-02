package sshserver

import (
	"fmt"
	"sync"
	"text/template"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/log"
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
	kbiName := config.DefaultKeyboardInteractiveName
	if cfg.KeyboardInteractiveName != nil {
		kbiName = *cfg.KeyboardInteractiveName
	}
	var kbiNameTemplate *template.Template
	if kbiName != "" {
		kbiNameTemplate, err = template.New("keyboardInteractiveName").Parse(kbiName)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the keyboardInteractiveName template (%w)", err)
		}
	}
	return &serverImpl{
		cfg:             cfg,
		handler:         handler,
		logger:          logger,
		kbiNameTemplate: kbiNameTemplate,
		wg:              &sync.WaitGroup{},
		lock:            &sync.Mutex{},
		listenSocket:    nil,
		hostKeys:        hostKeys,
		shutdownHandlers: &shutdownRegistry{
			lock:      &sync.Mutex{},
			callbacks: map[string]shutdownHandler{},
		},
	}, nil
}
