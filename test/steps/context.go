package steps

import (
	"github.com/janoszen/containerssh/log"
	"github.com/janoszen/containerssh/test/auth"
	"github.com/janoszen/containerssh/test/config"
	"github.com/janoszen/containerssh/test/ssh"
)

// region Context
type Scenario struct {
	Logger       log.Logger
	LogWriter    log.Writer
	AuthServer   *auth.MemoryAuthServer
	ConfigServer *config.MemoryConfigServer
	SshServer    *ssh.Server
}
