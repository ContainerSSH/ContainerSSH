package security //nolint:testpackage

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/stretchr/testify/assert"
)

func TestEnvRequest(t *testing.T) {
	session := &sessionHandler{
		config: config.SecurityConfig{
			Env: config.SecurityEnvConfig{
				Allow: []string{"ALLOW_ME"},
				Deny:  []string{"DENY_ME"},
			},
		},
		backend: &dummyBackend{},
		sshConnection: &sshConnectionHandler{
			lock: &sync.Mutex{},
		},
		logger: log.NewTestLogger(t),
	}

	session.config.Env.Mode = config.ExecutionPolicyEnable
	assert.NoError(t, session.OnEnvRequest(1, "ALLOW_ME", "bar"))
	assert.NoError(t, session.OnEnvRequest(2, "OTHER", "bar"))
	assert.Error(t, session.OnEnvRequest(3, "DENY_ME", "bar"))

	session.config.Env.Mode = config.ExecutionPolicyFilter
	assert.NoError(t, session.OnEnvRequest(4, "ALLOW_ME", "bar"))
	assert.Error(t, session.OnEnvRequest(5, "OTHER", "bar"))
	assert.Error(t, session.OnEnvRequest(6, "DENY_ME", "bar"))

	session.config.Env.Mode = config.ExecutionPolicyDisable
	assert.Error(t, session.OnEnvRequest(7, "ALLOW_ME", "bar"))
	assert.Error(t, session.OnEnvRequest(8, "OTHER", "bar"))
	assert.Error(t, session.OnEnvRequest(9, "DENY_ME", "bar"))
}

func TestPTYRequest(t *testing.T) {
	session := &sessionHandler{
		config:  config.SecurityConfig{},
		backend: &dummyBackend{},
		sshConnection: &sshConnectionHandler{
			lock: &sync.Mutex{},
		},
		logger: log.NewTestLogger(t),
	}

	session.config.TTY.Mode = config.ExecutionPolicyEnable
	assert.NoError(t, session.OnPtyRequest(1, "XTERM", 80, 25, 800, 600, []byte{}))

	session.config.TTY.Mode = config.ExecutionPolicyFilter
	assert.Error(t, session.OnPtyRequest(1, "XTERM", 80, 25, 800, 600, []byte{}))

	session.config.TTY.Mode = config.ExecutionPolicyDisable
	assert.Error(t, session.OnPtyRequest(1, "XTERM", 80, 25, 800, 600, []byte{}))
}

func TestCommand(t *testing.T) {
	backend := &dummyBackend{}
	session := &sessionHandler{
		config:  config.SecurityConfig{},
		backend: backend,
		sshConnection: &sshConnectionHandler{
			lock: &sync.Mutex{},
		},
		logger: log.NewTestLogger(t),
	}

	session.config.Command.Allow = []string{"/bin/bash"}
	session.config.Command.Mode = config.ExecutionPolicyDisable
	assert.Error(t, session.OnExecRequest(1, "/bin/bash"))

	session.config.Command.Mode = config.ExecutionPolicyFilter
	assert.NoError(t, session.OnExecRequest(1, "/bin/bash"))
	assert.Error(t, session.OnExecRequest(1, "/bin/sh"))

	session.config.Command.Mode = config.ExecutionPolicyEnable
	assert.NoError(t, session.OnExecRequest(1, "/bin/bash"))
	assert.NoError(t, session.OnExecRequest(1, "/bin/sh"))

	session.config.Shell.Mode = config.ExecutionPolicyEnable
	backend.commandsExecuted = []string{}
	backend.env = map[string]string{}
	assert.NoError(t, session.OnExecRequest(1, "/bin/bash"))
	assert.Equal(t, []string{"/bin/bash"}, backend.commandsExecuted)
	assert.Equal(t, map[string]string{}, backend.env)

	session.config.Shell.Mode = config.ExecutionPolicyEnable
	session.config.ForceCommand = "/bin/wrapper"
	backend.commandsExecuted = []string{}
	backend.env = map[string]string{}
	assert.NoError(t, session.OnExecRequest(1, "/bin/bash"))
	assert.Equal(t, []string{"/bin/wrapper"}, backend.commandsExecuted)
	assert.Equal(t, map[string]string{"SSH_ORIGINAL_COMMAND": "/bin/bash"}, backend.env)
}

func TestShell(t *testing.T) {
	backend := &dummyBackend{}
	session := &sessionHandler{
		config:  config.SecurityConfig{},
		backend: backend,
		sshConnection: &sshConnectionHandler{
			lock: &sync.Mutex{},
		},
		logger: log.NewTestLogger(t),
	}

	session.config.Shell.Mode = config.ExecutionPolicyDisable
	assert.Error(t, session.OnShell(1))

	session.config.Shell.Mode = config.ExecutionPolicyFilter
	assert.Error(t, session.OnShell(1))

	session.config.Shell.Mode = config.ExecutionPolicyEnable
	assert.NoError(t, session.OnShell(1))

	session.config.Shell.Mode = config.ExecutionPolicyEnable
	backend.commandsExecuted = []string{}
	backend.env = map[string]string{}
	assert.NoError(t, session.OnShell(1))
	assert.Equal(t, []string{"shell"}, backend.commandsExecuted)
	assert.Equal(t, map[string]string{}, backend.env)

	session.config.Shell.Mode = config.ExecutionPolicyEnable
	session.config.ForceCommand = "/bin/wrapper"
	backend.commandsExecuted = []string{}
	backend.env = map[string]string{}
	assert.NoError(t, session.OnShell(1))
	assert.Equal(t, []string{"/bin/wrapper"}, backend.commandsExecuted)
}

func TestSubsystem(t *testing.T) {
	backend := &dummyBackend{}
	session := &sessionHandler{
		config:  config.SecurityConfig{},
		backend: backend,
		sshConnection: &sshConnectionHandler{
			lock: &sync.Mutex{},
		},
		logger: log.NewTestLogger(t),
	}

	session.config.Subsystem.Mode = config.ExecutionPolicyDisable
	assert.Error(t, session.OnSubsystem(1, "sftp"))

	session.config.Subsystem.Mode = config.ExecutionPolicyFilter
	assert.Error(t, session.OnSubsystem(1, "sftp"))
	session.config.Subsystem.Allow = []string{"sftp"}
	assert.NoError(t, session.OnSubsystem(1, "sftp"))

	session.config.Subsystem.Mode = config.ExecutionPolicyEnable
	session.config.Subsystem.Allow = []string{}
	assert.NoError(t, session.OnSubsystem(1, "sftp"))
	session.config.Subsystem.Deny = []string{"sftp"}
	assert.Error(t, session.OnSubsystem(1, "sftp"))

	session.config.Subsystem.Mode = config.ExecutionPolicyEnable
	backend.commandsExecuted = []string{}
	session.config.Subsystem.Deny = []string{}
	backend.env = map[string]string{}
	assert.NoError(t, session.OnSubsystem(1, "sftp"))
	assert.Equal(t, []string{"sftp"}, backend.commandsExecuted)
	assert.Equal(t, map[string]string{}, backend.env)

	session.config.Subsystem.Mode = config.ExecutionPolicyEnable
	session.config.ForceCommand = "/bin/wrapper"
	backend.commandsExecuted = []string{}
	session.config.Subsystem.Deny = []string{}
	backend.env = map[string]string{}
	assert.NoError(t, session.OnSubsystem(1, "sftp"))
	assert.Equal(t, []string{"/bin/wrapper"}, backend.commandsExecuted)
	assert.Equal(t, map[string]string{"SSH_ORIGINAL_COMMAND": "sftp"}, backend.env)
}

// region Dummy backend
type dummyBackend struct {
	exit             chan struct{}
	env              map[string]string
	commandsExecuted []string
}

func (d *dummyBackend) OnClose() {
}

func (d *dummyBackend) OnShutdown(_ context.Context) {
}

func (d *dummyBackend) OnUnsupportedChannelRequest(_ uint64, _ string, _ []byte) {

}

func (d *dummyBackend) OnFailedDecodeChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
	_ error,
) {

}

func (d *dummyBackend) OnEnvRequest(_ uint64, name string, value string) error {
	if d.env != nil {
		d.env[name] = value
	}
	return nil
}

func (d *dummyBackend) OnPtyRequest(
	_ uint64,
	_ string,
	_ uint32,
	_ uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	return nil
}

func (d *dummyBackend) OnExecRequest(
	_ uint64,
	program string,
) error {
	d.commandsExecuted = append(d.commandsExecuted, program)
	return nil
}

func (d *dummyBackend) OnShell(
	_ uint64,
) error {
	d.commandsExecuted = append(d.commandsExecuted, "shell")

	go func() {
		if d.exit != nil {
			<-d.exit
		}
	}()
	return nil
}

func (d *dummyBackend) OnSubsystem(
	_ uint64,
	subsystem string,
) error {
	d.commandsExecuted = append(d.commandsExecuted, subsystem)

	go func() {
		if d.exit != nil {
			<-d.exit
		}
	}()
	return nil
}

func (d *dummyBackend) OnSignal(_ uint64, _ string) error {
	return nil
}

func (d *dummyBackend) OnWindow(_ uint64, _ uint32, _ uint32, _ uint32, _ uint32) error {
	return nil
}

func (s *dummyBackend) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

// endregion
