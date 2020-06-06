package backend

import (
	"fmt"
	"io"
)

type ShellOrSubsystem struct {
	Stdin  io.Writer
	Stdout io.Reader
	Stderr io.Reader
}

type Session interface {
	//Set an environment variable for the application. this can only be called before RequestShell or RequestSubsystem
	//have been called.
	SetEnv(name string, value string) error
	//Resize the application to fit the new window size.
	Resize(cols uint, rows uint) error
	//Request that a pseudoterminal be allocated when RequestShell or RequestSubsystem are called.
	SetPty() error
	//Request a shell. If a PTY has been requested before this command will start with a PTY. It will return a construct
	//with stdin, stdout and stderr. It is the callers responsibility to watch for the end of these streams and call
	//Close afterwards
	RequestShell() (*ShellOrSubsystem, error)
	//Request a subsystem. If a PTY has been requested before this command will start with a PTY. It will return a
	//construct with stdin, stdout and stderr. It is the callers responsibility to watch for the end of these streams
	//and call Close afterwards
	RequestSubsystem(subsystem string) (*ShellOrSubsystem, error)
	//Request the exit code of the program. Less than zero means that no exit code is available yet.
	GetExitCode() int32
	//Clean up the shell
	Close()
	//Send a signal to the container
	SendSignal(signal string) error
}

type Backend struct {
	Name          string
	CreateSession func(sessionId string, username string) (Session, error)
}

type Registry struct {
	backends    map[string]Backend
	backendKeys []string
}

func NewRegistry() *Registry {
	return &Registry{
		backends:    make(map[string]Backend),
		backendKeys: []string{},
	}
}

func (registry *Registry) Register(backend Backend) {
	registry.backends[backend.Name] = backend
	registry.backendKeys = append(registry.backendKeys, backend.Name)
}

func (registry *Registry) GetBackends() []string {
	return registry.backendKeys
}

func (registry *Registry) GetBackend(key string) (*Backend, error) {
	if backend, ok := registry.backends[key]; ok {
		return &backend, nil
	}
	return nil, fmt.Errorf("unknown backend (%s)", key)
}
