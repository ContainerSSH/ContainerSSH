package kubernetes

import (
	"context"

	"k8s.io/client-go/tools/remotecommand"
)

type pushSizeQueueImpl struct {
	resizeChan chan remotecommand.TerminalSize
}

func (s *pushSizeQueueImpl) Push(ctx context.Context, size remotecommand.TerminalSize) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.resizeChan <- size:
		return nil
	}
}

func (s *pushSizeQueueImpl) Next() *remotecommand.TerminalSize {
	size, ok := <-s.resizeChan
	if !ok {
		return nil
	}
	return &size
}

func (s *pushSizeQueueImpl) Stop() {
	close(s.resizeChan)
}
