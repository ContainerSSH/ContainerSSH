package dockerrun

import (
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mattn/go-shellwords"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

func (session *dockerRunSession) RequestProgram(program string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) error {
	if session.containerId != "" {
		return fmt.Errorf("cannot change request program after the container has started")
	}

	config := session.config

	if config.Config.DisableCommand && program != "" {
		return fmt.Errorf("command execution disabled, cannot run program: %s", program)
	}


	image := config.Config.ContainerConfig.Image
	_, err := reference.ParseNamed(config.Config.ContainerConfig.Image)
	if err != nil {
		if err == reference.ErrNameNotCanonical {
			if !strings.Contains(config.Config.ContainerConfig.Image, "/") {
				image = "docker.io/library/" + image
			} else {
				image = "docker.io/" + image
			}
		} else {
			return err
		}
	}

	pullReader, err := session.client.ImagePull(session.ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(pullReader)
	if err != nil {
		return err
	}
	err = pullReader.Close()
	if err != nil {
		return err
	}

	config.Config.ContainerConfig.AttachStdin = true
	config.Config.ContainerConfig.AttachStderr = true
	config.Config.ContainerConfig.AttachStdout = true
	config.Config.ContainerConfig.OpenStdin = true
	config.Config.ContainerConfig.StdinOnce = true
	config.Config.ContainerConfig.Tty = session.pty

	for key, value := range session.env {
		config.Config.ContainerConfig.Env = append(config.Config.ContainerConfig.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if !config.Config.DisableCommand && program != "" {
		programParts, err := shellwords.Parse(program)
		if err != nil {
			config.Config.ContainerConfig.Cmd = []string{"/bin/sh", "-c", program}
		} else {
			if strings.HasPrefix(programParts[0], "/") || strings.HasPrefix(programParts[0], "./") || strings.HasPrefix(programParts[0], "../") {
				config.Config.ContainerConfig.Cmd = programParts
			} else {
				config.Config.ContainerConfig.Cmd = []string{"/bin/sh", "-c", program}
			}
		}
	}

	body, err := session.client.ContainerCreate(
		session.ctx,
		&config.Config.ContainerConfig,
		&config.Config.HostConfig,
		&config.Config.NetworkConfig,
		config.Config.ContainerName,
	)
	if err != nil {
		return err
	}
	session.containerId = body.ID
	attachOptions := types.ContainerAttachOptions{
		Logs:   true,
		Stdin:  true,
		Stderr: true,
		Stdout: true,
		Stream: true,
	}
	attachResult, err := session.client.ContainerAttach(session.ctx, session.containerId, attachOptions)
	if err != nil {
		return err
	}

	startOptions := types.ContainerStartOptions{}
	err = session.client.ContainerStart(session.ctx, session.containerId, startOptions)
	if err != nil {
		return err
	}

	var once sync.Once
	if session.pty {
		go func() {
			_, _ = io.Copy(stdOut, attachResult.Reader)
			once.Do(done)
		}()
	} else {
		go func() {
			//Demultiplex Docker stream
			_, _ = stdcopy.StdCopy(stdOut, stdErr, attachResult.Reader)
			once.Do(done)
		}()
	}
	go func() {
		_, _ = io.Copy(attachResult.Conn, stdIn)
		once.Do(done)
	}()

	return nil
}
