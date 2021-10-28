package sshserver

import (
	"testing"

	"github.com/containerssh/libcontainerssh/log"
)

// ConformanceTestBackendFactory is a method to creating a network connection conformanceTestHandler for testing purposes.
type ConformanceTestBackendFactory = func(t *testing.T, logger log.Logger) (NetworkConnectionHandler, error)
