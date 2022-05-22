package unixutils

import (
	"github.com/mattn/go-shellwords"
)

// ParseCMD takes a shell command line and parses it into an execv-compatible slice.
func ParseCMD(cmd string) ([]string, error) {
	return shellwords.Parse(cmd)
}
