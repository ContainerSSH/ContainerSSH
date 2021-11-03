package test

import (
	"net"
	"sync"
	"testing"
)

var lock = &sync.Mutex{}

// GetNextPort returns the next free port a test service can bind to.
func GetNextPort(t *testing.T) int {
	// Note: the t parameter is requested so that later on a more advanced allocation algorithm can be implemented with
	// a proper cleanup.

	lock.Lock()
	defer lock.Unlock()

	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("Failed to allocate port for test %s (%v)", t.Name(), err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err = l.Close(); err != nil {
		t.Fatalf("Failed to close temporary listen socket on port %d for test %s (%v)", port, t.Name(), err)
	}

	t.Logf("Allocating port %d for test %s", port, t.Name())
	return port
}
