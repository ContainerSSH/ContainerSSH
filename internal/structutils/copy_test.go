package structutils_test

import (
	"testing"

    "go.containerssh.io/libcontainerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
)

type copyTest struct {
	Sub *copyTestSub
}

type copyTestSub struct {
	Text string
}

func TestCopy(t *testing.T) {
	data := new(copyTest)
	data.Sub = new(copyTestSub)
	data.Sub.Text = "Hello world!"

	newData := new(copyTest)
	err := structutils.Copy(newData, data)
	assert.Nil(t, err, "failed to copy struct (%v)", err)

	data.Sub.Text = "Hello world 2"

	assert.Equal(t, "Hello world!", newData.Sub.Text)
	assert.Equal(t, "Hello world 2", data.Sub.Text)
}
