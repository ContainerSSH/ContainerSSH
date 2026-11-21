package sshserver

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func conformanceShellCommand(t *testing.T, session TestClientSession, command string, expectResponse string) bool {
	if err := session.Type([]byte(fmt.Sprintf("%s\n", command))); err != nil {
		t.Error(err)
		return false
	}
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := session.WaitForStdout(
		timeout,
		[]byte(expectResponse),
	); err != nil {
		t.Error(err)
		return false
	}
	if !strings.Contains("exit", command) {
		session.ReadRemaining()
	}
	return true
}
