package steps

import (
	"github.com/containerssh/log"
	"github.com/containerssh/service"

	"github.com/containerssh/containerssh/test/auth"
	"github.com/containerssh/containerssh/test/config"
)

// region Context
type Scenario struct {
	Logger        log.Logger
	AuthServer    *auth.MemoryAuthServer
	ConfigServer  *config.MemoryConfigServer
	LoggerFactory log.LoggerFactory
	Lifecycle     service.Lifecycle
}
