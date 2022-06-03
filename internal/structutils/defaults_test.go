package structutils_test

import (
	"testing"

    "go.containerssh.io/libcontainerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
)

type defaultTest struct {
	Text     string `default:"Hello world!"`
	Decision bool   `default:"true"`
	Number   int    `default:"42"`
}

func TestDefaultsShouldSetDefaults(t *testing.T) {
	data := new(defaultTest)
	structutils.Defaults(data)
	assert.Equal(t, "Hello world!", data.Text)
	assert.Equal(t, true, data.Decision)
	assert.Equal(t, 42, data.Number)
}
