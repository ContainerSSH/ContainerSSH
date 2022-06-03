package sshserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

    "go.containerssh.io/libcontainerssh/log"
    messageCodes "go.containerssh.io/libcontainerssh/message"
	"golang.org/x/crypto/ssh"
)

type testClientSessionImpl struct {
	session  *ssh.Session
	stdin    *syncContextPipe
	stderr   *syncContextPipe
	stdout   *syncContextPipe
	exitCode int
	logger   log.Logger
	pty      bool
}

func (t *testClientSessionImpl) ReadRemaining() {
	t.logger.Debug(messageCodes.NewMessage(
		messageCodes.MTest,
		"Reading remaining bytes from stdout...",
	))
	for {
		if t.readOne() {
			return
		}
	}
}

func (t *testClientSessionImpl) readOne() (done bool) {
	data := make([]byte, 1024)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := t.stdout.ReadCtx(ctx, data)
	return err != nil
}

func (t *testClientSessionImpl) ReadRemainingStderr() {
	t.logger.Debug(messageCodes.NewMessage(
		messageCodes.MTest,
		"Reading remaining bytes from stderr...",
	))
	for {
		data := make([]byte, 1024)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := t.stderr.ReadCtx(ctx, data)
		if err != nil {
			return
		}
	}
}

func (t *testClientSessionImpl) Type(data []byte) error {
	t.logger.Debug(messageCodes.NewMessage(
		messageCodes.MTest,
		"Typing on stdin with sleep and read back: %s",
		data,
	))
	for _, b := range data {
		if _, err := t.Write([]byte{b}); err != nil {
			return err
		}
		readBack := make([]byte, 1)
		n, err := t.Read(readBack)
		if err != nil {
			return err
		}
		if n == 1 && readBack[0] == '\r' {
			n, err = t.Read(readBack)
			if err != nil {
				return err
			}
		}
		if n != 1 || b != readBack[0] {
			// Read the rest of the output so we get a useful message:
			for {
				readCtx, cancel := context.WithTimeout(context.Background(), time.Second)
				buf := make([]byte, 1024)
				n, err := t.ReadCtx(readCtx, buf)
				cancel()
				if err != nil {
					break
				}
				readBack = append(readBack, buf[:n]...)
			}
			return fmt.Errorf("failed to read back typed byte '%s' found: %s", []byte{b}, readBack)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Typing done."))
	return nil
}

func (t *testClientSessionImpl) Signal(signal string) error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Sending %s signal to process...", signal))
	return t.session.Signal(ssh.Signal(signal))
}

func (t *testClientSessionImpl) MustSignal(signal string) {
	err := t.Signal(signal)
	if err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) MustSetEnv(name string, value string) {
	if err := t.SetEnv(name, value); err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) MustWindow(cols int, rows int) {
	if err := t.Window(cols, rows); err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) MustRequestPTY(term string, cols int, rows int) {
	if err := t.RequestPTY(term, cols, rows); err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) MustShell() {
	if err := t.Shell(); err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) MustExec(program string) {
	if err := t.Exec(program); err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) MustSubsystem(name string) {
	if err := t.Subsystem(name); err != nil {
		panic(err)
	}
}

func (t *testClientSessionImpl) SetEnv(name string, value string) error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Setting env variable %s=%s...", name, value))
	return t.session.Setenv(name, value)
}

func (t *testClientSessionImpl) Window(cols int, rows int) error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Changing window to cols %d rows %d...", cols, rows))
	return t.session.WindowChange(rows, cols)
}

func (t *testClientSessionImpl) RequestPTY(term string, cols int, rows int) error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Requesting PTY for term %s cols %d rows %d...", term, cols, rows))
	if err := t.session.RequestPty(term, rows, cols, ssh.TerminalModes{}); err != nil {
		return err
	}
	t.pty = true
	return nil
}

func (t *testClientSessionImpl) Shell() error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Executing shell..."))
	if t.pty {
		t.session.Stderr = nil
	}
	return t.session.Shell()
}

func (t *testClientSessionImpl) Exec(program string) error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Executing program '%s'...", program))
	if t.pty {
		t.session.Stderr = nil
	}
	return t.session.Start(program)
}

func (t *testClientSessionImpl) Subsystem(name string) error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Requesting subsystem %s...", name))
	if t.pty {
		t.session.Stderr = nil
	}
	return t.session.RequestSubsystem(name)
}

func (t *testClientSessionImpl) Write(data []byte) (int, error) {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Writing to stdin: %s", data))
	return t.stdin.Write(data)
}

func (t *testClientSessionImpl) WriteCtx(ctx context.Context, data []byte) (int, error) {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Writing to stdin: %s", data))
	return t.stdin.WriteCtx(ctx, data)
}

func (t *testClientSessionImpl) Read(data []byte) (int, error) {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Reading %d bytes from stdout...", len(data)))
	return t.stdout.Read(data)
}

func (t *testClientSessionImpl) ReadCtx(ctx context.Context, data []byte) (int, error) {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Reading %d bytes from stdout...", len(data)))
	return t.stdout.ReadCtx(ctx, data)
}

func (t *testClientSessionImpl) WaitForStdout(ctx context.Context, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Waiting for the following string on stdout: %s", data))
	if len(data) == 0 {
		return nil
	}
	ringBuffer := make([]byte, len(data))
	bufIndex := 0
	for {
		buf := make([]byte, 1)
		n, err := t.stdout.ReadCtx(ctx, buf)
		if err != nil {
			return err
		}
		if n > 0 {
			if bufIndex == len(data) {
				ringBuffer = append(ringBuffer[1:], buf[0])
			} else {
				ringBuffer[bufIndex] = buf[0]
				bufIndex += n
			}
		}
		t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Ringbuffer currently contains the following %d bytes: %s", bufIndex, ringBuffer[:bufIndex]))
		if bytes.Equal(ringBuffer[:bufIndex], data) {
			return nil
		}
	}
}

func (t *testClientSessionImpl) Stderr() io.Reader {
	return t.stderr
}

func (t *testClientSessionImpl) Wait() error {
	t.logger.Debug(messageCodes.NewMessage(messageCodes.MTest, "Waiting for session to finish."))
	t.ReadRemaining()
	t.ReadRemainingStderr()
	err := t.session.Wait()
	if err != nil {
		exitErr := &ssh.ExitError{}
		if errors.As(err, &exitErr) {
			t.exitCode = exitErr.ExitStatus()
			return nil
		}
	} else {
		t.exitCode = 0
	}
	return err
}

func (t *testClientSessionImpl) ExitCode() int {
	return t.exitCode
}

func (t *testClientSessionImpl) Close() error {
	return t.session.Close()
}

func newSyncContextPipe() *syncContextPipe {
	return &syncContextPipe{
		make(chan byte),
		false,
		&sync.Mutex{},
	}
}
