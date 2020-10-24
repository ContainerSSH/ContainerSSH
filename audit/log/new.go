package log

import (
	"github.com/containerssh/containerssh/audit"
	"github.com/containerssh/containerssh/log"
)

func New(logger log.Logger) audit.Plugin {
	return &Plugin{
		logger: logger,
	}
}
