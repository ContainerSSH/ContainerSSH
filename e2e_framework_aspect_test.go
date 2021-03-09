package containerssh_test

import (
	"github.com/containerssh/configuration"
	"github.com/containerssh/log"
)

// TestingAspect describes one aspect that is being tested. For example, an aspect may be the status of audit logs being
// enabled or disabled.
//
// Each aspect can introduce multiple test cases that are run in combination to filter out any wiring bugs.
type TestingAspect interface {
	String() string
	Factors() []TestingFactor
}

// TestingFactor is a single factor within a testing aspect.
type TestingFactor interface {
	Aspect() TestingAspect
	String() string
	ModifyConfiguration(config *configuration.AppConfig) error
	StartBackingServices(config configuration.AppConfig, logger log.Logger) error
	GetSteps(config configuration.AppConfig, logger log.Logger) []Step
	StopBackingServices(config configuration.AppConfig, logger log.Logger) error
}
