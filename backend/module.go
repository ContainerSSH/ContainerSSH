package backend

import (
	"fmt"
	"github.com/janoszen/containerssh/config"
	"github.com/janoszen/containerssh/log"
	"io"
)

type ShellOrSubsystem struct {
}

type Session interface {
	//Set an environment variable for the application. this can only be called before RequestProgram or RequestSubsystem
	//have been called.
	SetEnv(name string, value string) error
	//Resize the application to fit the new window size.
	Resize(cols uint, rows uint) error
	//Request that a pseudo terminal be allocated when RequestProgram or RequestSubsystem are called.
	SetPty() error
	//Request a shell. If a PTY has been requested before this command will start with a PTY. It will return a construct
	//with stdin, stdout and stderr. It is the callers responsibility to watch for the end of these streams and call
	//Close afterwards
	RequestProgram(program string, stdin io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) error
	//Request the execution of a subsystem. It works similar to RequestProgram, except that it takes a
	//subsystem name instead of a program.
	RequestSubsystem(subsystem string, stdin io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) error
	//Request the exit code of the program. Less than zero means that no exit code is available yet.
	GetExitCode() int32
	//Clean up the shell
	Close()
	//Send a signal to the container
	SendSignal(signal string) error
}

type Backend struct {
	Name          config.BackendName
	CreateSession func(sessionId string, username string, appConfig *config.AppConfig, logger log.Logger) (Session, error)
}

type Registry struct {
	backends    map[config.BackendName]Backend
	backendKeys []config.BackendName
}

func NewRegistry() *Registry {
	return &Registry{
		backends:    make(map[config.BackendName]Backend),
		backendKeys: []config.BackendName{},
	}
}

func (registry *Registry) Register(backend Backend) {
	registry.backends[backend.Name] = backend
	registry.backendKeys = append(registry.backendKeys, backend.Name)
}

func (registry *Registry) GetBackends() []config.BackendName {
	return registry.backendKeys
}

func (registry *Registry) GetBackend(key config.BackendName) (*Backend, error) {
	if backend, ok := registry.backends[key]; ok {
		return &backend, nil
	}
	return nil, fmt.Errorf("unknown backend (%s)", key)
}
