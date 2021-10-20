package unixutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/containerssh/unixutils"
)

func TestParseCMD(t *testing.T) {
	args, err := unixutils.ParseCMD("/bin/sh -c 'echo \"Hello world!\"'")
	assert.Nil(t, err, "failed to parse CMD (%v)", err)
	assert.Equal(t, []string{"/bin/sh", "-c", "echo \"Hello world!\""}, args)
}
