package containerssh_test

import (
	"os"
	"os/exec"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.containerssh.io/containerssh"
)

func TestVersionFlag(t *testing.T) {
	if os.Getenv("CONTAINERSSH_TEST_VERSION_FLAG") == "1" {
		// Child process: Main parses os.Args and exits.
		os.Args = []string{"containerssh", "--version"}
		containerssh.Main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestVersionFlag$")
	cmd.Env = append(os.Environ(), "CONTAINERSSH_TEST_VERSION_FLAG=1")
	// There is no config file in this directory, and --version should not need one.
	cmd.Dir = t.TempDir()
	output, err := cmd.Output()
	assert.NoError(t, err)

	info, ok := debug.ReadBuildInfo()
	assert.True(t, ok)
	assert.Equal(t, info.Main.Version+"\n", string(output))
}
