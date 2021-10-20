package log_test

import (
	"testing"

	"github.com/containerssh/containerssh/log"
)

func TestMain(m *testing.M)  {
	log.RunTests(m)
}