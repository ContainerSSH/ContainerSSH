package test

import (
	"net"
	"sync"
	"testing"
)

var lock = &sync.Mutex{}

// GetNextPort returns the next free port a test service can bind to.
func GetNextPort(t *testing.T, purpose string) int {
	t.Helper()

	lock.Lock()
	defer lock.Unlock()

	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("Failed to allocate port for test %s for %s (%v)", t.Name(), purpose, err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	if err = l.Close(); err != nil {
		t.Fatalf(
			"Failed to close temporary listen socket on port %d for test %s for %s (%v)",
			port,
			t.Name(),
			purpose,
			err,
		)
	}

	t.Logf("Allocating port %d for test %s for %s", port, t.Name(), purpose)
	return port
}
