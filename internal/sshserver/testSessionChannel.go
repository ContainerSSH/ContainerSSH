package sshserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/containerssh/libcontainerssh/internal/unixutils"
)

type testSessionChannel struct {
	AbstractSessionChannelHandler

	session SessionChannel
	env     map[string]string
	pty     bool
	rows    uint32
	columns uint32
	running bool
	term    bool
}

func (t *testSessionChannel) OnEnvRequest(_ uint64, name string, value string) error {
	if t.running {
		return errors.New("cannot set env variable to an already running program")
	}
	t.env[name] = value
	return nil
}

func (t *testSessionChannel) OnPtyRequest(
	_ uint64,
	term string,
	columns uint32,
	rows uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	if t.running {
		return errors.New("cannot set PTY for an already running program")
	}
	t.pty = true
	t.env["TERM"] = term
	t.rows = rows
	t.columns = columns
	return nil
}

func (t *testSessionChannel) OnExecRequest(
	_ uint64,
	program string,
) error {
	if t.running {
		return errors.New("program already running")
	}
	argv, err := unixutils.ParseCMD(program)
	if err != nil {
		return err
	}
	t.running = true
	go func() {
		err := t.run(argv, t.session.Stdout(), t.session.Stderr(), true)
		if err != nil {
			t.session.ExitStatus(1)
		} else {
			t.session.ExitStatus(0)
		}
		_ = t.session.Close()
	}()
	return nil
}

func (t *testSessionChannel) OnShell(
	_ uint64,
) error {
	if t.running {
		return errors.New("program already running")
	}
	t.running = true
	go func() {
		for {
			if t.pty {
				_, err := t.session.Stdout().Write([]byte("> "))
				if err != nil {
					t.session.ExitStatus(1)
					_ = t.session.Close()
					return
				}
			}
			command, done := t.readCommand(t.session.Stdin(), t.session.Stdout())
			if done {
				return
			}
			argv, err := unixutils.ParseCMD(command)
			if err != nil {
				_, _ = t.session.Stderr().Write([]byte(err.Error()))
				t.session.ExitStatus(1)
				_ = t.session.Close()
				return
			}
			var stderr io.Writer
			if t.pty {
				// If the terminal is interactive everything goes to stdout.
				stderr = t.session.Stdout()
			} else {
				stderr = t.session.Stderr()
			}
			if argv[0] == "exit" {
				t.session.ExitStatus(0)
				_ = t.session.Close()
			}
			if err := t.run(argv, t.session.Stdout(), stderr, false); err != nil {
				t.session.ExitStatus(1)
				_ = t.session.Close()
			}
		}
	}()
	return nil
}

func (t *testSessionChannel) readCommand(
	stdin io.Reader,
	stdout io.Writer,
) (
	string,
	bool,
) {
	cmd := bytes.Buffer{}
	for {
		b := make([]byte, 1)
		n, err := stdin.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.session.ExitStatus(0)
				_ = t.session.Close()
				return "", true
			} else {
				t.session.ExitStatus(1)
				_ = t.session.Close()
				return "", true
			}
		}
		if n == 0 {
			t.session.ExitStatus(0)
			_ = t.session.Close()
			return "", true
		}
		if t.term {
			t.session.ExitStatus(0)
			_ = t.session.Close()
			return "", true
		}
		if _, err := stdout.Write(b); err != nil {
			if errors.Is(err, io.EOF) {
				t.session.ExitStatus(0)
			} else {
				t.session.ExitStatus(1)
			}
			_ = t.session.Close()
			return "", true
		}
		cmd.Write(b)
		if b[0] == '\n' {
			break
		}
	}
	command := strings.TrimSpace(cmd.String())
	return command, false
}

func (t *testSessionChannel) run(argv []string, stdout io.Writer, stderr io.Writer, exitWithError bool) (err error) {
	switch argv[0] {
	case "echo":
		_, err = stdout.Write([]byte(strings.Join(argv[1:], " ") + "\n"))
	case "tput":
		if len(argv) > 2 || (argv[1] != "cols" && argv[1] != "rows") {
			_, err = stderr.Write([]byte("Usage: tput [rows|cols]"))
			if exitWithError {
				return fmt.Errorf("usage: tput [rows|cols]")
			}
		} else if !t.pty {
			_, err = stderr.Write([]byte("Stdout is not a TTY"))
			if exitWithError {
				return fmt.Errorf("usage: tput [rows|cols]")
			}
		} else {
			switch argv[1] {
			case "cols":
				_, err = stdout.Write([]byte(fmt.Sprintf("%d\n", t.columns)))
			case "rows":
				_, err = stdout.Write([]byte(fmt.Sprintf("%d\n", t.rows)))
			}
		}
	default:
		_, err = stderr.Write([]byte(fmt.Sprintf("unknown program: %s", argv[0])))
		if exitWithError {
			return fmt.Errorf("unknown program: %s", argv[0])
		}
	}
	return err
}

func (t *testSessionChannel) OnSignal(
	_ uint64,
	signal string,
) error {
	if !t.running {
		return errors.New("program not running")
	}
	if signal != "TERM" {
		return fmt.Errorf("signal type not supported")
	}

	t.session.ExitStatus(0)
	_ = t.session.Close()

	return nil
}

func (t *testSessionChannel) OnWindow(
	_ uint64,
	columns uint32,
	rows uint32,
	_ uint32,
	_ uint32,
) error {
	if !t.running {
		return errors.New("program not running")
	}
	if !t.pty {
		return errors.New("not a PTY session")
	}
	t.rows = rows
	t.columns = columns
	return nil
}

func (s *testSessionChannel) OnX11Request(
	requestID uint64,
	singleConnection bool,
	protocol string,
	cookie string,
	screen uint32,
	reverseHandler ReverseForward,
) error {
	return fmt.Errorf("Unimplemented")
}

func (t *testSessionChannel) OnShutdown(_ context.Context) {
	if t.running {
		_ = t.session.Close()
	}
}
