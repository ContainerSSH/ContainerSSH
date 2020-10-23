package steps

import (
	"github.com/containerssh/containerssh/log"
	"github.com/containerssh/containerssh/test/auth"
	"github.com/containerssh/containerssh/test/config"
	"github.com/containerssh/containerssh/test/ssh"
)

// region Context
type Scenario struct {
	Logger       log.Logger
	LogWriter    log.Writer
	AuthServer   *auth.MemoryAuthServer
	ConfigServer *config.MemoryConfigServer
	SshServer    *ssh.Server
}
