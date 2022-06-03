package test_test

import (
	"fmt"
	"net"
	"testing"

    "go.containerssh.io/libcontainerssh/internal/test"
)

func TestPortAllocation(t *testing.T) {
	port := test.GetNextPort(t, "test")

	listen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to listen on pre-allocated port %d (%v).", port, err)
	}
	_ = listen.Close()
}
