package sshserver_test

import (
	"testing"

	"github.com/containerssh/libcontainerssh/log"
)

func TestMain(m *testing.M) {
	log.RunTests(m)
}
