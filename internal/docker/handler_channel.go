package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

    "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/internal/unixutils"
    "go.containerssh.io/libcontainerssh/message"
)

type channelHandler struct {
	sshserver.AbstractSessionChannelHandler

	channelID         uint64
	networkHandler    *networkHandler
	connectionHandler *sshConnectionHandler
	username          string
	env               map[string]string
	pty               bool
	columns           uint32
	rows              uint32
	exitSent          bool
	x11               bool
	exec              dockerExecution
	session           sshserver.SessionChannel
}

func (c *channelHandler) OnEnvRequest(_ uint64, name string, value string) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return message.UserMessage(message.EDockerProgramAlreadyRunning, "program already running", "program already running")
	}
	c.env[name] = value
	return nil
}

func (c *channelHandler) OnPtyRequest(
	_ uint64,
	term string,
	columns uint32,
	rows uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return message.UserMessage(message.EDockerProgramAlreadyRunning, "program already running", "program already running")
	}
	c.env["TERM"] = term
	c.rows = rows
	c.columns = columns
	c.pty = true
	return nil
}

func (c *channelHandler) parseProgram(program string) []string {
	programParts, err := unixutils.ParseCMD(program)
	if err != nil {
		return []string{"/bin/sh", "-c", program}
	} else {
		if strings.HasPrefix(programParts[0], "/") || strings.HasPrefix(
			programParts[0],
			"./",
		) || strings.HasPrefix(programParts[0], "../") {
			return programParts
		} else {
			return []string{"/bin/sh", "-c", program}
		}
	}
}

func (c *channelHandler) run(
	ctx context.Context,
	program []string,
) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return message.UserMessage(message.EDockerProgramAlreadyRunning, "program already running", "program already running")
	}

	var err error
	switch c.networkHandler.config.Execution.Mode {
	case config.DockerExecutionModeConnection:
		err = c.handleExecModeConnection(ctx, program)
	case config.DockerExecutionModeSession:
		err = c.handleExecModeSession(ctx, program)
	default:
		err = message.UserMessage(
			message.EDockerConfigError,
			"cannot run program",
			"invalid execution mode: %s",
			c.networkHandler.config.Execution.Mode,
		)
	}
	if err != nil {
		return err
	}

	c.exec.run(
		c.session.Stdin(),
		c.session.Stdout(),
		c.session.Stderr(),
		c.session.CloseWrite,
		func(exitStatus int) {
			c.session.ExitStatus(uint32(exitStatus))
			if err := c.session.Close(); err != nil && !errors.Is(err, io.EOF) {
				c.networkHandler.logger.Debug(
					message.Wrap(
						err,
						message.EDockerFailedOutputCloseWriting,
						"failed to close session",
					))
			}
		},
	)

	return nil
}

func (c *channelHandler) handleExecModeConnection(
	ctx context.Context,
	program []string,
) error {
	exec, err := c.networkHandler.container.createExec(ctx, program, c.env, c.pty)
	if err != nil {
		return err
	}
	c.exec = exec
	if c.pty {
		err = c.exec.resize(ctx, uint(c.rows), uint(c.columns))
		if err != nil {
			c.networkHandler.logger.Debug(err)
		}
	}
	return nil
}

func (c *channelHandler) handleExecModeSession(
	ctx context.Context,
	program []string,
) error {
	cnt, err := c.networkHandler.dockerClient.createContainer(
		ctx,
		c.networkHandler.labels,
		c.env,
		&c.pty,
		program,
	)
	if err != nil {
		return err
	}
	removeContainer := func() {
		ctx, cancelFunc := context.WithTimeout(
			context.Background(), c.networkHandler.config.Timeouts.ContainerStop,
		)
		defer cancelFunc()
		_ = cnt.remove(ctx)
	}
	c.exec, err = cnt.attach(ctx)
	if err != nil {
		removeContainer()
		return err
	}
	if err := cnt.start(ctx); err != nil {
		removeContainer()
		return err
	}
	if c.pty {
		err := c.exec.resize(ctx, uint(c.rows), uint(c.columns))
		if err != nil {
			removeContainer()
			return err
		}
	}
	return nil
}

func (c *channelHandler) OnExecRequest(
	_ uint64,
	program string,
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()
	return c.run(
		startContext,
		c.parseProgram(program),
	)
}

func (c *channelHandler) OnShell(
	_ uint64,
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()

	return c.run(startContext, c.getDefaultShell())
}

func (c *channelHandler) getDefaultShell() []string {
	return c.networkHandler.config.Execution.ShellCommand
}

func (c *channelHandler) OnSubsystem(
	_ uint64,
	subsystem string,
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()

	if binary, ok := c.networkHandler.config.Execution.Subsystems[subsystem]; ok {
		return c.run(startContext, []string{binary})
	}
	return message.UserMessage(message.EDockerSubsystemNotSupported, "subsystem not supported", "the specified subsystem is not supported (%s)", subsystem)
}

func (c *channelHandler) OnSignal(_ uint64, signal string) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return message.UserMessage(
			message.EDockerProgramNotRunning,
			"Cannot send signal, program is not running.",
			"Cannot send signal, program is not running.",
		)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.Signal)
	defer cancelFunc()

	return c.exec.signal(ctx, signal)
}

func (c *channelHandler) OnWindow(_ uint64, columns uint32, rows uint32, _ uint32, _ uint32) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return message.UserMessage(
			message.EDockerProgramNotRunning,
			"Cannot resize window, program is not running.",
			"Cannot resize window, program is not running.",
		)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.Window)
	defer cancelFunc()

	return c.exec.resize(ctx, uint(rows), uint(columns))
}

func (c *channelHandler) OnX11Request(
	requestID uint64,
	singleConnection bool,
	proto string,
	cookie string,
	screen uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	if c.x11 {
		return fmt.Errorf("X11 forwarding already setup for this channel")
	}

	c.env["DISPLAY"] = "localhost:10"

	err := c.connectionHandler.agentForward.NewX11Forwarding(
		c.connectionHandler.setupAgent,
		c.networkHandler.logger,
		singleConnection,
		proto,
		cookie,
		screen,
		reverseHandler,
	)
	if err != nil {
		return err
	}
	c.x11 = true

	return nil
}

func (c *channelHandler) OnClose() {
	if c.exec != nil {
		c.exec.kill()
	}
	if c.x11 {
		err := c.connectionHandler.agentForward.CloseX11Forwarding()
		if err != nil {
			c.networkHandler.logger.Info(fmt.Errorf("Failed to close X11 forwarding (%w)", err))
		}
	}
	container := c.networkHandler.container
	if container != nil && c.networkHandler.config.Execution.Mode == config.DockerExecutionModeSession {
		ctx, cancel := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.ContainerStop)
		defer cancel()
		_ = container.remove(ctx)
	}
}

func (c *channelHandler) OnShutdown(shutdownContext context.Context) {
	if c.exec != nil {
		c.exec.term(shutdownContext)
		// We wait for the program to exit. This is not needed in session or connection mode, but
		// later we will need to support persistent containers.
		select {
		case <-shutdownContext.Done():
			c.exec.kill()
		case <-c.exec.done():
		}
	}
}
