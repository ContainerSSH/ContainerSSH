package log

import (
	"io"
	"sync"

    "go.containerssh.io/libcontainerssh/config"
)

// newStdoutWriter creates a log writer that writes to the stdout (io.Writer) in the specified format.
func newStdoutWriter(stdout io.Writer, format config.LogFormat) Writer {
	return &stdoutWriter{
		fileHandleWriter: newFileHandleWriter(stdout, format, &sync.Mutex{}),
	}
}

// stdoutWriter inherits the write method from fileHandleWriter and writes to the stdout.
type stdoutWriter struct {
	*fileHandleWriter
}
