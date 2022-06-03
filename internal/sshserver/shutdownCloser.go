package sshserver

import (
	"context"
	"io"
)

type shutdownCloser struct {
	closer io.Closer
}

func (s *shutdownCloser) OnShutdown(_ context.Context) {
	_ = s.closer.Close()
}