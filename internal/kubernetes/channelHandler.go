package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/sshserver"
	"go.containerssh.io/containerssh/internal/unixutils"
	"go.containerssh.io/containerssh/message"
)

type channelHandler struct {
	channelID         uint64
	networkHandler    *networkHandler
	connectionHandler *sshConnectionHandler
	username          string
	env               map[string]string
	files             map[string][]byte
	pty               bool
	x11               bool
	columns           uint32
	rows              uint32
	exec              kubernetesExecution
	session           sshserver.SessionChannel
	pod               kubernetesPod
}

func (c *channelHandler) OnUnsupportedChannelRequest(_ uint64, _ string, _ []byte) {
}

func (c *channelHandler) OnFailedDecodeChannelRequest(
	_ uint64,
	_ string,
	_ []byte,
	_ error,
) {
}

func (c *channelHandler) OnEnvRequest(_ uint64, name string, value string) error {
	if c.exec != nil {
		return message.UserMessage(message.EKubernetesProgramAlreadyRunning, "program already running", "program already running")
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
	if c.exec != nil {
		return message.UserMessage(message.EKubernetesProgramAlreadyRunning, "program already running", "program already running")
	}
	c.env["TERM"] = term
	c.pty = true
	c.columns = columns
	c.rows = rows
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

	var err error
	switch c.networkHandler.config.Pod.Mode {
	case config.KubernetesExecutionModePersistent:
		c.pod, err = c.handleExecModePersistent(ctx, program)
	case config.KubernetesExecutionModeConnection:
		err = c.handleExecModeConnection(ctx, program)
	case config.KubernetesExecutionModeSession:
		c.pod, err = c.handleExecModeSession(ctx, program)
	default:
		// This should never happen due to validation.
		return fmt.Errorf("invalid execution mode: %s", c.networkHandler.config.Pod.Mode)
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
						message.EKubernetesFailedOutputCloseWriting,
						"failed to close session",
					))
			}
		},
	)

	if c.pty {
		err = c.exec.resize(ctx, uint(c.rows), uint(c.columns))
		if err != nil {
			c.networkHandler.logger.Debug(message.Wrap(err,
				message.EKubernetesFailedResize, "Failed to set initial terminal size"))
		}
	}

	return nil
}

func (c *channelHandler) handleExecModePersistent(ctx context.Context, program []string) (kubernetesPod, error) {
	pod, err := c.networkHandler.cli.findPod(
		ctx,
		c.networkHandler.config.Pod.Metadata.Name,
		c.networkHandler.config.Pod.Metadata.Namespace,
	)
	if err != nil {
		if strings.Contains(err.Error(), message.EKubernetesPodNotFound) {
			if c.networkHandler.config.Pod.CreateMissingPods {
				c.networkHandler.logger.Debug(
					message.NewMessage(
						message.EKubernetesPodNotFound,
						fmt.Sprintf("Pod %s not found in namespace %s, spawning pod",
							c.networkHandler.config.Pod.Metadata.Name,
							c.networkHandler.config.Pod.Metadata.Namespace),
					),
				)
				pod, err := c.networkHandler.cli.createPod(
					ctx,
					c.networkHandler.labels,
					c.networkHandler.annotations,
					c.env,
					&c.pty,
					program,
				)
				if err != nil {
					return nil, err
				}
				c.exec, err = pod.attach(ctx)
				if err != nil {
					return nil, err
				}
				return pod, nil
			}
		}
		return nil, err
	}
	c.pod = pod
	// c.networkHandler.logger.Debug(c.env)
	exec, err := c.networkHandler.pod.createExec(ctx, program, c.env, c.pty)
	if err == nil {
		c.exec = exec
		return nil, err
	}
	return nil, nil
}

func (c *channelHandler) handleExecModeConnection(
	ctx context.Context,
	program []string,
) error {
	exec, err := c.networkHandler.pod.createExec(ctx, program, c.env, c.pty)
	if err != nil {
		return err
	}
	c.exec = exec
	return nil
}

func (c *channelHandler) handleExecModeSession(
	ctx context.Context,
	program []string,
) (kubernetesPod, error) {
	pod, err := c.networkHandler.cli.createPod(
		ctx,
		c.networkHandler.labels,
		c.networkHandler.annotations,
		c.env,
		&c.pty,
		program,
	)
	if err != nil {
		return nil, err
	}

	for path, content := range c.files {
		ctx, cancelFunc := context.WithTimeout(
			context.Background(),
			c.networkHandler.config.Timeouts.CommandStart,
		)
		c.networkHandler.logger.Debug(message.NewMessage(
			message.MKubernetesFileModification,
			"Writing to file %s",
			path,
		))
		defer cancelFunc()
		err := pod.writeFile(ctx, path, content)
		if err != nil {
			c.networkHandler.logger.Warning(message.Wrap(
				err,
				message.EKubernetesFileModificationFailed,
				"Failed to write to %s",
				path,
			))
		}
	}

	c.exec, err = pod.attach(ctx)
	if err != nil {
		c.removePod(pod)
		return nil, err
	}
	return pod, nil
}

func (c *channelHandler) removePod(pod kubernetesPod) {
	ctx, cancelFunc := context.WithTimeout(
		context.Background(), c.networkHandler.config.Timeouts.PodStop,
	)
	defer cancelFunc()
	_ = pod.remove(ctx)
}

func (c *channelHandler) OnExecRequest(
	_ uint64,
	program string,
) error {
	startContext, cancelFunc := context.WithTimeout(
		context.Background(),
		c.networkHandler.config.Timeouts.CommandStart,
	)
	defer cancelFunc()

	return c.run(startContext, c.parseProgram(program))
}

func (c *channelHandler) OnShell(
	_ uint64,
) error {
	startContext, cancelFunc := context.WithTimeout(
		context.Background(),
		c.networkHandler.config.Timeouts.CommandStart,
	)
	defer cancelFunc()

	return c.run(startContext, c.networkHandler.config.Pod.ShellCommand)
}

func (c *channelHandler) OnSubsystem(
	_ uint64,
	subsystem string,
) error {
	startContext, cancelFunc := context.WithTimeout(
		context.Background(),
		c.networkHandler.config.Timeouts.CommandStart,
	)
	defer cancelFunc()

	if binary, ok := c.networkHandler.config.Pod.Subsystems[subsystem]; ok {
		return c.run(startContext, []string{binary})
	}
	return message.UserMessage(message.EKubernetesSubsystemNotSupported, "subsystem not supported", "the specified subsystem is not supported (%s)", subsystem)
}

func (c *channelHandler) OnSignal(_ uint64, signal string) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return message.UserMessage(
			message.EKubernetesProgramNotRunning,
			"Cannot send signal, program is not running.",
			"Cannot send signal, program is not running.",
		)
	}
	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		c.networkHandler.config.Timeouts.Signal,
	)
	defer cancelFunc()

	return c.exec.signal(ctx, signal)
}

func (c *channelHandler) OnWindow(_ uint64, columns uint32, rows uint32, _ uint32, _ uint32) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return message.UserMessage(
			message.EKubernetesProgramNotRunning,
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

	c.networkHandler.logger.Debug("Setting X11 vars")
	c.env["DISPLAY"] = "localhost:10"
	c.networkHandler.logger.Warning("Kubernetes: Starting reverse forward")

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
			c.networkHandler.logger.Info(fmt.Errorf("failed to close X11 forwarding (%w)", err))
		}
	}
	pod := c.networkHandler.pod
	if pod != nil && c.networkHandler.config.Pod.Mode == config.KubernetesExecutionModeSession {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			c.networkHandler.config.Timeouts.PodStop,
		)
		defer cancel()
		_ = pod.remove(ctx)
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
