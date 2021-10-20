package sshserver

import (
	"github.com/containerssh/containerssh/log"
)

// ConformanceTestBackendFactory is a method to creating a network connection conformanceTestHandler for testing purposes.
type ConformanceTestBackendFactory = func(logger log.Logger) (NetworkConnectionHandler, error)
