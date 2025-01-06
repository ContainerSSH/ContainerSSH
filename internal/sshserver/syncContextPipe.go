package sshserver

import (
	"context"
	"io"
	"sync"
)

// syncContextPipe is a pipe that is able to handle timeouts via contexts.
type syncContextPipe struct {
	byteChannel chan byte
	closed      bool
	lock        *sync.Mutex
}

func (c *syncContextPipe) Write(data []byte) (int, error) {
	return c.WriteCtx(context.Background(), data)
}

func (c *syncContextPipe) Read(data []byte) (int, error) {
	return c.ReadCtx(context.Background(), data)
}

func (c *syncContextPipe) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return io.EOF
	}
	c.closed = true
	return nil
}

func (c *syncContextPipe) WriteCtx(ctx context.Context, data []byte) (int, error) {
	for _, b := range data {
		select {
		case c.byteChannel <- b:
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
	return len(data), nil
}

func (c *syncContextPipe) ReadCtx(ctx context.Context, data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	i := 0
	var b byte
	select {
	case b = <-c.byteChannel:
		data[i] = b
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	i++
	for {
		if i == len(data) {
			return i, nil
		}
		select {
		case b = <-c.byteChannel:
			data[i] = b
		case <-ctx.Done():
			return i, ctx.Err()
		default:
			return i, nil
		}
	}
}
