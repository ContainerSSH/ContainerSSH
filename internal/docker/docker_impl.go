package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/metrics"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type dockerV20ClientFactory struct {
	// backendFailuresMetric counts the failed requests to the backend.
	backendFailuresMetric metrics.SimpleCounter
	// backendRequestsMetric counts the requests to the backend.
	backendRequestsMetric metrics.SimpleCounter
}

func (f *dockerV20ClientFactory) getDockerClient(ctx context.Context, config config.DockerConfig) (*client.Client, error) {
	httpClient, err := getHTTPClient(config)
	if err != nil {
		return nil, err
	}
	cli, err := client.NewClientWithOpts(
		client.WithHost(config.Connection.Host),
		client.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(ctx)
	return cli, nil
}

func (f *dockerV20ClientFactory) get(ctx context.Context, config config.DockerConfig, logger log.Logger) (dockerClient, error) {
	if config.Execution.Launch.ContainerConfig == nil || config.Execution.Launch.ContainerConfig.Image == "" {
		return nil, message.NewMessage(message.EDockerConfigError, "no image name specified")
	}

	dockerClient, err := f.getDockerClient(ctx, config)
	if err != nil {
		return nil, err
	}

	return &dockerV20Client{
		config:       config,
		dockerClient: dockerClient,
		logger:       logger,

		backendFailuresMetric: f.backendFailuresMetric,
		backendRequestsMetric: f.backendRequestsMetric,
	}, nil
}

type dockerV20Client struct {
	config       config.DockerConfig
	dockerClient *client.Client
	logger       log.Logger

	// backendFailuresMetric counts the failed requests to the backend.
	backendFailuresMetric metrics.SimpleCounter
	// backendRequestsMetric counts the requests to the backend.
	backendRequestsMetric metrics.SimpleCounter
}

func (d *dockerV20Client) getImageName() string {
	return d.config.Execution.Launch.ContainerConfig.Image
}

func (d *dockerV20Client) hasImage(ctx context.Context) (bool, error) {
	image := d.config.Execution.Launch.ContainerConfig.Image
	d.logger.Debug(message.NewMessage(message.MDockerImageList, "Checking if image %s exists locally...", image))
	var lastError error
loop:
	for {
		d.backendRequestsMetric.Increment()
		_, _, lastError := d.dockerClient.ImageInspectWithRaw(ctx, image)
		if lastError != nil {
			if client.IsErrNotFound(lastError) {
				return false, nil
			}
			d.backendFailuresMetric.Increment()
			d.logger.Notice(
				message.Wrap(lastError,
					message.EDockerFailedImageList, "failed to list images, retrying in 10 seconds"))
		} else {
			return true, nil
		}
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	return false, message.Wrap(lastError, message.EDockerFailedImageList, "failed to list images, giving up")
}

func (d *dockerV20Client) pullImage(ctx context.Context) error {
	image, err := getCanonicalImageName(d.config.Execution.Launch.ContainerConfig.Image)
	if err != nil {
		return err
	}

	d.logger.Debug(message.NewMessage(message.MDockerImagePull, "Pulling image %s...", image))
	var lastError error
loop:
	for {
		var pullReader io.ReadCloser
		d.backendRequestsMetric.Increment()
		pullReader, lastError = d.dockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
		if lastError == nil {
			_, lastError = ioutil.ReadAll(pullReader)
			if lastError == nil {
				lastError = pullReader.Close()
				if lastError == nil {
					return nil
				}
			}
		}
		d.backendFailuresMetric.Increment()
		if pullReader != nil {
			_ = pullReader.Close()
		}
		d.logger.Notice(
			message.Wrap(
				lastError,
				message.EDockerFailedImagePull,
				"failed to pull image %s, retrying in 10 seconds",
				image,
			))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err = message.WrapUser(
		lastError,
		message.EDockerFailedImagePull,
		UserMessageInitializeSSHSession,
		"failed to pull image %s, retrying in 10 seconds",
		image,
	)
	d.logger.Debug(err)
	return err
}

func (d *dockerV20Client) createContainer(
	ctx context.Context,
	labels map[string]string,
	env map[string]string,
	tty *bool,
	cmd []string,
) (dockerContainer, error) {
	logger := d.logger
	logger.Debug(message.NewMessage(message.MDockerContainerCreate, "Creating container..."))
	containerConfig := d.config.Execution.Launch.ContainerConfig
	newConfig, err := d.createConfig(containerConfig, labels, env, tty, cmd)
	if err != nil {
		return nil, err
	}

	var lastError error
loop:
	for {
		var body container.ContainerCreateCreatedBody
		d.backendRequestsMetric.Increment()
		body, lastError = d.dockerClient.ContainerCreate(
			ctx,
			newConfig,
			d.config.Execution.Launch.HostConfig,
			d.config.Execution.Launch.NetworkConfig,
			d.config.Execution.Launch.Platform,
			d.config.Execution.Launch.ContainerName,
		)
		if lastError == nil {
			return &dockerV20Container{
				config:                d.config,
				containerID:           body.ID,
				dockerClient:          d.dockerClient,
				logger:                logger.WithLabel("containerId", body.ID),
				tty:                   newConfig.Tty,
				backendRequestsMetric: d.backendRequestsMetric,
				backendFailuresMetric: d.backendFailuresMetric,
				lock:                  &sync.Mutex{},
				wg:                    &sync.WaitGroup{},
				removeLock:            &sync.Mutex{},
			}, nil
		}
		d.backendFailuresMetric.Increment()
		logger.Debug(
			message.Wrap(lastError,
				message.EDockerFailedContainerCreate, "failed to create container, retrying in 10 seconds"))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err = message.WrapUser(
		lastError,
		message.EDockerFailedContainerCreate,
		UserMessageInitializeSSHSession,
		"failed to create container, giving up",
	)
	logger.Error(err)
	return nil, err
}

func (d *dockerV20Client) createConfig(
	containerConfig *container.Config,
	labels map[string]string,
	env map[string]string,
	tty *bool,
	cmd []string,
) (*container.Config, error) {
	newConfig := &container.Config{}
	if containerConfig != nil {
		if err := structutils.Copy(newConfig, containerConfig); err != nil {
			return nil, err
		}
	}
	if newConfig.Labels == nil {
		newConfig.Labels = map[string]string{}
	}
	newConfig.Cmd = d.config.Execution.IdleCommand
	for k, v := range labels {
		newConfig.Labels[k] = v
	}

	newConfig.Env = append(newConfig.Env, createEnv(env)...)
	if tty != nil {
		newConfig.Tty = *tty
		newConfig.AttachStdin = true
		newConfig.AttachStdout = true
		newConfig.AttachStderr = true
		newConfig.OpenStdin = true
		newConfig.StdinOnce = true
		newConfig.Cmd = cmd
	}
	return newConfig, nil
}

type dockerV20Container struct {
	config                config.DockerConfig
	containerID           string
	logger                log.Logger
	dockerClient          *client.Client
	tty                   bool
	backendRequestsMetric metrics.SimpleCounter
	backendFailuresMetric metrics.SimpleCounter
	lock                  *sync.Mutex
	wg                    *sync.WaitGroup
	shuttingDown          bool
	shutdown              bool
	removeLock            *sync.Mutex
}

func (d *dockerV20Container) attach(ctx context.Context) (dockerExecution, error) {
	d.logger.Debug(message.NewMessage(message.MDockerContainerAttach, "attaching to container..."))
	var attachResult types.HijackedResponse
	var lastError error
loop:
	for {
		d.backendRequestsMetric.Increment()
		attachResult, lastError = d.dockerClient.ContainerAttach(
			ctx,
			d.containerID,
			types.ContainerAttachOptions{
				Stream: true,
				Stdin:  true,
				Stdout: true,
				Stderr: true,
				Logs:   true,
			},
		)
		if lastError == nil {
			return &dockerV20Exec{
				container:    d,
				execID:       "",
				dockerClient: d.dockerClient,
				logger:       d.logger,
				attachResult: attachResult,
				tty:          d.tty,
				pid:          1,
				doneChan:     make(chan struct{}),
				lock:         &sync.Mutex{},
			}, nil
		}
		d.backendFailuresMetric.Increment()
		d.logger.Warning(
			message.Wrap(lastError,
				message.EDockerFailedContainerAttach, "failed to attach to exec, retrying in 10 seconds"))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.WrapUser(
		lastError,
		message.EDockerFailedContainerAttach,
		UserMessageInitializeSSHSession,
		"failed to attach to exec, giving up",
	)
	d.logger.Error(err)
	return nil, err
}

func (d *dockerV20Container) start(ctx context.Context) error {
	d.logger.Debug(message.NewMessage(message.MDockerContainerStart, "Starting container..."))
	var lastError error
loop:
	for {
		d.backendRequestsMetric.Increment()
		lastError = d.dockerClient.ContainerStart(
			ctx,
			d.containerID,
			types.ContainerStartOptions{},
		)
		if lastError == nil {
			return nil
		}
		d.backendFailuresMetric.Increment()
		d.logger.Debug(
			message.Wrap(lastError,
				message.EDockerFailedContainerStart, "failed to start container, retrying in 10 seconds"))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.WrapUser(
		lastError,
		message.EDockerFailedContainerStart,
		UserMessageInitializeSSHSession,
		"failed to start container, giving up",
	)
	d.logger.Error(err)
	return err
}

func (d *dockerV20Container) writeFile(path string, content []byte) error {
	if d.config.Execution.DisableAgent {
		return message.NewMessage(
			message.EDockerWriteFileFailed,
			"The ContainerSSH guest agent is disabled. Failed to write file %s in the container",
			path,
		)
	}

	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		d.config.Timeouts.CommandStart,
	)
	d.logger.Debug(message.NewMessage(
		message.MDockerFileWrite,
		"Writing to file %s",
		path,
	))
	defer cancelFunc()

	exec, err := d.createExec(
		ctx,
		[]string{d.config.Execution.AgentPath, "write-file", path},
		nil,
		false,
	)
	if err != nil {
		return message.Wrap(
			err,
			message.EDockerWriteFileFailed,
			"Failed to write file inside container",
		)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdin := bytes.NewReader(content)
	exitCode := -1
	done := make(chan int)
	exec.run(
		stdin,
		&stdout,
		&stderr,
		func() error {
			return nil
		},
		func(exitStatus int) {
			exitCode = exitStatus
			close(done)
		},
	)
	<-done
	if exitCode != 0 {
		return message.NewMessage(
			message.EDockerWriteFileFailed,
			"Agent exited with status %d error output: %s",
			exitCode,
			stderr.String(),
		)
	}

	return nil
}

func (d *dockerV20Container) remove(ctx context.Context) error {
	d.removeLock.Lock()
	defer d.removeLock.Unlock()
	if d.shuttingDown {
		return nil
	}

	d.lock.Lock()
	d.shuttingDown = true
	d.lock.Unlock()
	d.wg.Wait()
	d.lock.Lock()
	d.shutdown = true
	d.lock.Unlock()

	d.logger.Debug(message.NewMessage(message.MDockerContainerRemove, "Removing container..."))
	var lastError error
loop:
	for {
		d.backendRequestsMetric.Increment()
		_, lastError = d.dockerClient.ContainerInspect(ctx, d.containerID)
		if lastError != nil && client.IsErrNotFound(lastError) {
			return nil
		}

		if lastError == nil {
			d.backendRequestsMetric.Increment()
			lastError = d.dockerClient.ContainerRemove(
				ctx, d.containerID, types.ContainerRemoveOptions{
					Force: true,
				},
			)
			if lastError == nil {
				d.logger.Debug(message.NewMessage(message.MDockerContainerRemoveSuccessful, "Container removed."))
				return nil
			}
		}
		d.backendFailuresMetric.Increment()
		d.logger.Debug(
			message.Wrap(
				lastError,
				message.EDockerFailedContainerRemove,
				"failed to remove container on disconnect, retrying in 10 seconds",
			))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := message.Wrap(
		lastError,
		message.EDockerFailedContainerRemove,
		"failed to remove container on disconnect, giving up",
	)
	d.logger.Debug(
		err,
	)
	return err
}

func (d *dockerV20Container) createExec(
	ctx context.Context,
	program []string,
	env map[string]string,
	tty bool,
) (dockerExecution, error) {
	d.lock.Lock()
	if d.shuttingDown {
		return nil, message.UserMessage(
			message.EDockerShuttingDown,
			"Server is shutting down",
			"Refusing new Docker execution because the container is shutting down.",
		)
	}
	d.wg.Add(1)
	d.lock.Unlock()

	return d.lockedCreateExec(ctx, program, env, tty)
}

func (d *dockerV20Container) lockedCreateExec(
	ctx context.Context,
	program []string,
	env map[string]string,
	tty bool,
) (dockerExecution, error) {
	d.logger.Debug(message.NewMessage(message.MDockerExec, "Creating and attaching to container exec..."))
	execConfig := d.createExecConfig(env, tty, program)
	execID, err := d.realCreateExec(ctx, execConfig)
	if err != nil {
		d.wg.Done()
		return nil, err
	}

	attachResult, err := d.attachExec(ctx, execID, execConfig)
	if err != nil {
		d.wg.Done()
		return nil, err
	}

	pid := 0
	return &dockerV20Exec{
		container:    d,
		execID:       execID,
		dockerClient: d.dockerClient,
		logger:       d.logger,
		attachResult: attachResult,
		tty:          tty,
		pid:          pid,
		doneChan:     make(chan struct{}),
		lock:         &sync.Mutex{},
	}, nil
}

func (d *dockerV20Container) realCreateExec(ctx context.Context, execConfig types.ExecConfig) (string, error) {
	d.logger.Debug(message.NewMessage(message.MDockerExecCreate, "Creating exec..."))
	var lastError error
loop:
	for {
		var response types.IDResponse
		d.backendRequestsMetric.Increment()
		response, lastError = d.dockerClient.ContainerExecCreate(
			ctx,
			d.containerID,
			execConfig,
		)
		if lastError == nil {
			return response.ID, nil
		}
		d.backendFailuresMetric.Increment()
		if isPermanentError(lastError) {
			return "", message.Wrap(lastError, message.EDockerFailedExecCreate, "failed to create exec, permanent error")
		}
		d.logger.Debug(
			message.Wrap(lastError,
				message.EDockerFailedExecCreate, "failed to create exec, retrying in 10 seconds"))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.Wrap(lastError, message.EDockerFailedExecCreate, "failed to create exec, giving up")
	d.logger.Error(err)
	return "", err
}

func (d *dockerV20Container) createExecConfig(env map[string]string, tty bool, program []string) types.ExecConfig {
	dockerEnv := createEnv(env)
	if !d.config.Execution.DisableAgent {
		agentPrefix := []string{
			d.config.Execution.AgentPath,
			"console",
			"--pid",
			"--",
		}
		program = append(agentPrefix, program...)
	}
	execConfig := types.ExecConfig{
		Tty:          tty,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          dockerEnv,
		Cmd:          program,
	}
	return execConfig
}

func createEnv(env map[string]string) []string {
	dockerEnv := make([]string, len(env))
	i := 0
	for k, v := range env {
		dockerEnv[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return dockerEnv
}

func (d *dockerV20Container) attachExec(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error) {
	d.logger.Debug(message.NewMessage(message.MDockerExecAttach, "Attaching exec..."))
	var attachResult types.HijackedResponse
	var lastError error
loop:
	for {
		d.backendRequestsMetric.Increment()
		attachResult, lastError = d.dockerClient.ContainerExecAttach(
			ctx,
			execID,
			types.ExecStartCheck{
				Detach: false,
				Tty:    config.Tty,
			},
		)
		if lastError == nil {
			return attachResult, nil
		}
		d.backendFailuresMetric.Increment()
		if isPermanentError(lastError) {
			err := message.Wrap(lastError, message.EDockerFailedExecAttach, "failed to attach to exec, permanent error")
			d.logger.Debug(err)
			return types.HijackedResponse{}, err
		}
		d.logger.Debug(
			message.Wrap(lastError,
				message.EDockerFailedExecAttach, "failed to attach to exec, retrying in 10 seconds"))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.Wrap(lastError, message.EDockerFailedExecAttach, "failed to attach to exec, giving up")
	d.logger.Debug(err)
	return types.HijackedResponse{}, err
}

type dockerV20Exec struct {
	container    *dockerV20Container
	execID       string
	dockerClient *client.Client
	logger       log.Logger
	attachResult types.HijackedResponse
	tty          bool
	pid          int
	doneChan     chan struct{}
	lock         *sync.Mutex
}

func (d *dockerV20Exec) term(ctx context.Context) {
	select {
	case <-d.done():
		return
	default:
	}
	_ = d.signal(ctx, "TERM")
}

func (d *dockerV20Exec) kill() {
	select {
	case <-d.done():
		return
	default:
	}
	if d.execID != "" {
		_ = d.signal(context.Background(), "KILL")
	}
}

func (d *dockerV20Exec) done() <-chan struct{} {
	return d.doneChan
}

func (d *dockerV20Exec) signal(ctx context.Context, sig string) error {
	if d.container.config.Execution.DisableAgent {
		err := message.UserMessage(
			message.EDockerCannotSendSignalNoAgent,
			"Cannot send signal to process.",
			"Cannot send signal %s to process because the ContainerSSH agent is disabled",
			sig,
		).Label("signal", sig)
		d.logger.Debug(err)
		return err
	}
	d.lock.Lock()
	if d.pid <= 0 {
		d.lock.Unlock()
		return message.UserMessage(message.EDockerFailedSignalNoPID, "Cannot send signal to process", "could not send signal to exec, process ID not found")
	}
	if d.pid == 1 {
		d.lock.Unlock()
		return d.sendSignalToContainer(ctx, sig)
	}
	if d.container.shutdown {
		err := message.UserMessage(
			message.EDockerFailedExecSignal,
			"Cannot send signal to process.",
			"Not sending signal to process, container is already shutting down.",
		).Label("signal", sig)
		d.logger.Debug(err)
		return err
	}
	d.container.wg.Add(1)
	pid := d.pid
	d.lock.Unlock()
	if pid < 1 {
		err := message.UserMessage(
			message.EDockerFailedExecSignal,
			"Cannot send signal to process.",
			"No process ID recorded, not sending signal.",
		).Label("signal", sig)
		d.logger.Debug(err)
		return err
	}
	d.logger.Debug(
		message.NewMessage(
			message.MDockerExecSignal,
			"Using the exec facility to send signal %s to pid %d...",
			sig,
			pid,
		).Label("signal", sig))
	err := d.realSendSignal(ctx, sig, pid)
	if err != nil {
		d.logger.Debug(
			message.Wrap(
				err,
				message.EDockerFailedExecSignal,
				"Cannot send %s signal to container %s pid %d",
				sig, d.container.containerID, pid,
			).Label("signal", sig),
		)
	} else {
		d.logger.Debug(
			message.NewMessage(
				message.MDockerExecSignalSuccessful,
				"Sent %s signal to container %s pid %d",
				sig, d.container.containerID, pid,
			).Label("signal", sig),
		)
	}
	return err
}

func (d *dockerV20Exec) realSendSignal(ctx context.Context, sig string, pid int) error {
	exec, err := d.container.lockedCreateExec(
		ctx, []string{
			d.container.config.Execution.AgentPath,
			"signal",
			"--pid",
			strconv.Itoa(pid),
			"--signal",
			sig,
		}, map[string]string{}, false,
	)
	if err != nil {
		return err
	}
	var stdoutBytes bytes.Buffer
	var stderrBytes bytes.Buffer
	stdin, stdinWriter := io.Pipe()
	done := make(chan struct{})
	exec.run(
		stdin, &stdoutBytes, &stderrBytes, func() error {
			return nil
		}, func(exitStatus int) {
			if exitStatus != 0 {
				err = fmt.Errorf("signal program exited with status %d", exitStatus)
			}
			done <- struct{}{}
		},
	)
	<-done
	_ = stdinWriter.Close()
	return err
}

func (d *dockerV20Exec) sendSignalToContainer(ctx context.Context, sig string) error {
	d.logger.Debug(
		message.NewMessage(
			message.MDockerContainerSignal,
			"Sending the %s signal to container...",
			sig,
		).Label("signal", sig))
	var lastError error
loop:
	for {
		lastError = d.dockerClient.ContainerKill(ctx, d.container.containerID, sig)
		if lastError == nil {
			return nil
		}
		if isPermanentError(lastError) {
			err := message.WrapUser(
				lastError,
				message.EDockerFailedContainerSignal,
				"Cannot send signal to process.",
				"Cannot send %s signal to container %s, permanent error",
				sig,
				d.container.containerID,
			).Label("signal", sig)
			d.logger.Debug(err)
			return err
		}
		d.logger.Debug(
			message.Wrap(
				lastError,
				message.EDockerFailedContainerSignal,
				"Cannot send %s signal to container %s, retrying in 10 seconds",
				sig,
				d.container.containerID,
			).Label("signal", sig),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.Wrap(
		lastError,
		message.EDockerFailedContainerSignal,
		"Cannot send %s signal to container %s, retrying in 10 seconds",
		sig,
		d.container.containerID,
	).Label("signal", sig)
	d.logger.Debug(err)
	return err
}

func (d *dockerV20Exec) resize(ctx context.Context, height uint, width uint) error {
	d.logger.Debug(
		message.NewMessage(message.MDockerResizing, "Resizing window to %dx%d", width, height).
			Label("width", width).
			Label("height", height))
	var lastError error
loop:
	for {
		resizeOptions := types.ResizeOptions{
			Height: height,
			Width:  width,
		}
		if d.execID != "" {
			lastError = d.dockerClient.ContainerExecResize(
				ctx, d.execID, resizeOptions,
			)
		} else {
			lastError = d.dockerClient.ContainerResize(
				ctx, d.container.containerID, resizeOptions,
			)
		}
		if lastError == nil {
			return nil
		}
		if isPermanentError(lastError) {
			err := message.WrapUser(
				lastError,
				message.EDockerFailedResize,
				"Cannot resize window.",
				"cannot resize window, permanent error",
			).Label("height", height).Label("width", width)
			d.logger.Debug(err)
			return err
		}
		d.logger.Debug(
			message.Wrap(
				lastError,
				message.EDockerFailedResize,
				"cannot resize window, permanent error",
			).Label("height", height).Label("width", width))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.WrapUser(
		lastError,
		message.EDockerFailedResize,
		"Cannot resize window.",
		"cannot resize window, diving up",
	).Label("height", height).Label("width", width)
	d.logger.Debug(err)
	return err
}

func (d *dockerV20Exec) readBytesFromReader(source io.Reader, bytes uint) ([]byte, error) {
	finalBuffer := make([]byte, bytes)
	readIndex := uint(0)
	for {
		buf := make([]byte, bytes-readIndex)
		readBytes, err := source.Read(buf)
		copy(finalBuffer[readIndex:readBytes], buf[:readBytes])
		readIndex = readIndex + uint(readBytes)
		if err != nil {
			return finalBuffer[:readIndex], err
		}
		if readIndex == bytes {
			return finalBuffer, nil
		}
	}
}

func (d *dockerV20Exec) run(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	writeClose func() error,
	onExit func(exitStatus int),
) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.container.config.Execution.Mode == config.DockerExecutionModeConnection && !d.container.config.Execution.DisableAgent {
		if err := d.readPIDFromStdout(stdout); err != nil {
			d.logger.Error(
				message.Wrap(
					err,
					message.EDockerFailedPIDRead,
					"cannot read PID from container",
				))
			onExit(137)
			d.container.wg.Done()
			return
		}
	}
	once := &sync.Once{}
	exitFunc := func() {
		d.finished(onExit)
	}
	go d.processOutput(once, exitFunc, stdout, stderr, writeClose)
	go d.processInput(stdin, once, exitFunc)
}

func (d *dockerV20Exec) processInput(stdin io.Reader, once *sync.Once, exitFunc func()) {
	func() {
		defer once.Do(exitFunc)
		_, err := io.Copy(d.attachResult.Conn, stdin)
		if err != nil && !errors.Is(err, io.EOF) {
			d.logger.Debug(
				message.Wrap(
					err,
					message.EDockerFailedInputStream,
					"failed to stream input",
				),
			)
		}
		if err := d.attachResult.CloseWrite(); err != nil {
			d.logger.Debug(
				message.Wrap(
					err,
					message.EDockerFailedInputCloseWriting,
					"failed to close Docker attach for writing",
				),
			)
		}
	}()
}

func (d *dockerV20Exec) processOutput(
	once *sync.Once,
	exitFunc func(),
	stdout io.Writer,
	stderr io.Writer,
	writeClose func() error,
) {
	func() {
		defer once.Do(exitFunc)
		var err error
		if d.tty {
			_, err = io.Copy(stdout, d.attachResult.Reader)
		} else {
			_, err = stdcopy.StdCopy(stdout, stderr, d.attachResult.Reader)
		}
		if err != nil && !errors.Is(err, io.EOF) {
			d.logger.Error(
				message.Wrap(
					err,
					message.EDockerFailedOutputStream,
					"Cannot read PID from container",
				),
			)
		}
		if err := writeClose(); err != nil {
			d.logger.Debug(
				message.Wrap(
					err,
					message.EDockerFailedOutputCloseWriting,
					"failed to close SSH channel for writing",
				),
			)
		}
	}()
}

func (d *dockerV20Exec) readPIDFromStdout(stdout io.Writer) error {
	// Read PID from execution
	var pidBytes []byte
	var err error
	if d.tty {
		pidBytes, err = d.readBytesFromReader(d.attachResult.Reader, 4)
		if err != nil {
			return err
		}
	} else {
		// Read a single frame from the Docker socket to get the PID.
		// See https://docs.docker.com/engine/api/v1.41/#operation/ContainerAttach
		var headerBuffer []byte
		headerBuffer, err = d.readBytesFromReader(d.attachResult.Reader, 8)
		if err != nil {
			return message.Wrap(err, message.EDockerFailedAgentRead, "failed to read from ContainerSSH agent")
		}
		stream := headerBuffer[0]
		if stream > 2 {
			return fmt.Errorf("invalid stream type received from Docker daemon: %d", stream)
		}
		frameLength := binary.BigEndian.Uint32(headerBuffer[4:])
		frameData, err := d.readBytesFromReader(d.attachResult.Reader, uint(frameLength))
		if err != nil {
			return fmt.Errorf("failed to read pid from ContainerSSH agent (%w)", err)
		}
		if frameLength < 4 {
			return message.NewMessage(
				message.EDockerFailedAgentRead,
				"not enough testdata received (%d bytes) from Docker daemon while trying to read pid from ContainerSSH agent",
				frameLength,
			)
		}
		switch stream {
		case 0:
			fallthrough
		case 1:
			pidBytes = frameData[:4]
			if frameLength > 4 {
				if _, err := stdout.Write(frameData[4:]); err != nil {
					return fmt.Errorf("failed to write remaining frame testdata to stdout (%w)", err)
				}
			}
		case 2:
			return fmt.Errorf(
				"unexpected testdata on stderr when trying to read pid from ContainerSSH agent",
			)
		}
	}
	d.pid = int(binary.LittleEndian.Uint32(pidBytes))
	return nil
}

func (d *dockerV20Exec) finished(onExit func(exitStatus int)) {
	d.lock.Lock()
	if d.pid == -1 {
		d.lock.Unlock()
		return
	}
	d.pid = -1
	if d.execID != "" {
		defer d.container.wg.Done()
	}
	d.lock.Unlock()
	close(d.doneChan)
	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()
	var lastError error
loop:
	for {
		if d.execID != "" {
			if lastError = d.execInspect(ctx, onExit); lastError == nil {
				return
			}
		} else {
			if err := d.stopContainer(ctx); err != nil {
				onExit(137)
				return
			}

			if lastError = d.containerInspect(ctx, onExit); lastError == nil {
				return
			}
		}
		if isPermanentError(lastError) {
			err := message.Wrap(lastError,
				message.EDockerFetchingExitCodeFailed, "Failed to fetch exit code, permanent error")
			d.logger.Error(err)
			return
		}
		d.logger.Debug(
			message.Wrap(lastError,
				message.EDockerFetchingExitCodeFailed, "Failed to fetch exit code, retrying in 10 seconds"),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := message.Wrap(lastError, message.EDockerFetchingExitCodeFailed, "Failed to fetch exit code, giving up")
	d.logger.Error(err)
}

func (d *dockerV20Exec) containerInspect(
	ctx context.Context,
	onExit func(exitStatus int),
) (lastError error) {
	var inspectResult types.ContainerJSON

	inspectResult, lastError = d.dockerClient.ContainerInspect(ctx, d.container.containerID)
	if lastError == nil {
		if inspectResult.State.Running {
			lastError = message.NewMessage(message.EDockerStillRunning, "container still running")
		} else if inspectResult.State.Restarting {
			lastError = message.NewMessage(message.EDockerContainerRestarting, "container is restarting")
		} else if inspectResult.State.ExitCode < 0 {
			lastError = message.NewMessage(message.EDockerNegativeExitCode, "negative exit code: %d", inspectResult.State.ExitCode)
		} else {
			onExit(inspectResult.State.ExitCode)
			return nil
		}
	}
	return lastError
}

func (d *dockerV20Exec) execInspect(ctx context.Context, onExit func(exitStatus int)) (lastError error) {
	var inspectResult types.ContainerExecInspect
	inspectResult, lastError = d.dockerClient.ContainerExecInspect(ctx, d.execID)
	if lastError == nil {
		if inspectResult.Running {
			lastError = message.NewMessage(message.EDockerStillRunning, "Program still running")
		} else if inspectResult.ExitCode < 0 {
			lastError = message.NewMessage(
				message.EDockerNegativeExitCode,
				"Negative exit code: %d",
				inspectResult.ExitCode,
			).Label("exitCode", inspectResult.ExitCode)
		} else {
			err := message.NewMessage(message.MDockerExitCode, "Program exited with %d", inspectResult.ExitCode)
			d.logger.Debug(err)

			onExit(inspectResult.ExitCode)
			return nil
		}
	}
	return lastError
}

func (d *dockerV20Exec) stopContainer(ctx context.Context) error {
	d.logger.Debug(message.NewMessage(message.MDockerContainerStop, "Stopping container..."))
	var lastError error
loop:
	for {
		var inspectResult types.ContainerJSON
		d.container.backendRequestsMetric.Increment()
		inspectResult, lastError = d.dockerClient.ContainerInspect(ctx, d.container.containerID)
		if lastError == nil {
			if inspectResult.State.Status == "stopped" {
				return nil
			}
			lastError = d.dockerClient.ContainerStop(
				ctx,
				d.container.containerID,
				&d.container.config.Timeouts.ContainerStop)
			if lastError == nil {
				return nil
			}
		} else {
			d.container.backendFailuresMetric.Increment()
		}
		d.logger.Debug(
			message.Wrap(lastError, message.EDockerContainerStopFailed, "failed to stop container, retrying in 10 seconds"),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := message.Wrap(lastError, message.EDockerContainerStopFailed, "failed to stop container, giving up")
	d.logger.Error(err)
	return err
}

func isPermanentError(err error) bool {
	return client.IsErrNotFound(err) ||
		client.IsErrNotImplemented(err) ||
		client.IsErrPluginPermissionDenied(err) ||
		client.IsErrUnauthorized(err)
}
