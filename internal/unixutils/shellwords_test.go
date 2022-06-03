package unixutils_test

import (
	"testing"

    "go.containerssh.io/libcontainerssh/internal/unixutils"
	"github.com/stretchr/testify/assert"
)

func TestParseCMD(t *testing.T) {
	args, err := unixutils.ParseCMD("/bin/sh -c 'echo \"Hello world!\"'")
	assert.Nil(t, err, "failed to parse CMD (%v)", err)
	assert.Equal(t, []string{"/bin/sh", "-c", "echo \"Hello world!\""}, args)
}
