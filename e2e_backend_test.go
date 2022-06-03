package libcontainerssh_test

import (
	"fmt"
	"testing"

    "go.containerssh.io/libcontainerssh/config"
)

// TestBackends tests all possible backends for basic functionality.
func TestBackends(t *testing.T) {
	for _, backend := range config.BackendValues() {
		t.Run(
			fmt.Sprintf("backend=%s", backend),
			func(t *testing.T) {
				testBackend(NewT(t), backend)
			},
		)
	}
}

// testBackend tests a single backend.
func testBackend(t T, backend config.Backend) {
	t.ConfigureBackend(backend)
	t.StartContainerSSH()
	t.LoginViaSSH()

	t.Run("CommandExecution", func(t T) {
		t.StartSessionChannel()
		t.RequestCommandExecution("echo 'Hello world!' && sleep 3")
		t.AssertStdoutHas("Hello world!")
		t.CloseChannel()
	})

	t.Run("Shell", func(t T) {
		t.StartSessionChannel()
		t.RequestShell()
		t.SendStdin("echo 'Hello world!' && sleep 3\n")
		t.AssertStdoutHas("Hello world!")
		t.SendStdin("exit\n")
		t.CloseChannel()
	})
}
