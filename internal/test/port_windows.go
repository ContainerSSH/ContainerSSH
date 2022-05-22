//go:build windows || plan9
// +build windows plan9

package test

import (
	"syscall"
)

func socketControl(_ string, _ string, _ syscall.RawConn) error {
	return nil
}
