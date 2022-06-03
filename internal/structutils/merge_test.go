package structutils_test

import (
	"testing"

    "go.containerssh.io/libcontainerssh/internal/structutils"
	"github.com/stretchr/testify/assert"
)

type mergeTest struct {
	Sub *mergeTestSub
}

type mergeTestSub struct {
	Text     string
	Bool     bool
	StrSlice []string
}

func TestMerge(t *testing.T) {
	data := new(mergeTest)
	newData := new(mergeTest)
	newData.Sub = new(mergeTestSub)
	newData.Sub.Text = "Hello world!"

	err := structutils.Merge(data, newData)
	assert.Nil(t, err, "failed to merge testdata (%v)", err)

	assert.Equal(t, "Hello world!", data.Sub.Text)
}

func TestMergeExisting(t *testing.T) {
	data := new(mergeTest)
	data.Sub = new(mergeTestSub)
	data.Sub.Text = "Foo"
	newData := new(mergeTest)
	newData.Sub = new(mergeTestSub)
	newData.Sub.Text = "Hello world!"

	err := structutils.Merge(data, newData)
	assert.Nil(t, err, "failed to merge testdata (%v)", err)

	assert.Equal(t, "Hello world!", data.Sub.Text)
}

func TestMergeExistingEmpty(t *testing.T) {
	data := new(mergeTest)
	data.Sub = new(mergeTestSub)
	data.Sub.Text = "Foo"
	newData := new(mergeTest)

	err := structutils.Merge(data, newData)
	assert.Nil(t, err, "failed to merge testdata (%v)", err)

	assert.Equal(t, "Foo", data.Sub.Text)
}

func TestMergeExistingEmpty2(t *testing.T) {
	data := new(mergeTest)
	data.Sub = new(mergeTestSub)
	data.Sub.Text = "Foo"
	newData := new(mergeTest)
	newData.Sub = new(mergeTestSub)

	err := structutils.Merge(data, newData)
	assert.Nil(t, err, "failed to merge testdata (%v)", err)

	assert.Equal(t, "Foo", data.Sub.Text)
}

func TestMergeBool(t *testing.T) {
	data := new(mergeTest)
	data.Sub = new(mergeTestSub)
	data.Sub.Bool = false
	newData := new(mergeTest)
	newData.Sub = new(mergeTestSub)
	newData.Sub.Bool = true

	err := structutils.Merge(data, newData)
	assert.Nil(t, err, "failed to merge testdata (%v)", err)

	assert.Equal(t, true, data.Sub.Bool)
}

func TestMergeStringSlice(t *testing.T) {
	data := new(mergeTest)
	data.Sub = new(mergeTestSub)
	data.Sub.StrSlice = []string{"foo"}
	newData := new(mergeTest)
	newData.Sub = new(mergeTestSub)
	newData.Sub.StrSlice = []string{"bar"}

	err := structutils.Merge(data, newData)
	assert.Nil(t, err, "failed to merge testdata (%v)", err)

	assert.Equal(t, []string{"bar"}, data.Sub.StrSlice)
}
