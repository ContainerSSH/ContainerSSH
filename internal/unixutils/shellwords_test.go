package unixutils_test

import (
	"testing"

    "go.containerssh.io/containerssh/internal/unixutils"
	"github.com/stretchr/testify/assert"
)

func TestParseCMD(t *testing.T) {
	args, err := unixutils.ParseCMD("/bin/sh -c 'echo \"Hello world!\"'")
	assert.Nil(t, err, "failed to parse CMD (%v)", err)
	assert.Equal(t, []string{"/bin/sh", "-c", "echo \"Hello world!\""}, args)
}
