package dockerrun

import (
	"containerssh/backend"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/mattn/go-shellwords"
	"io/ioutil"
	"strings"
)

func (session *dockerRunSession) RequestProgram(program string) (*backend.ShellOrSubsystem, error) {
	if session.containerId != "" {
		return nil, fmt.Errorf("cannot change request shell after the container has started")
	}

	config := session.config

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
			return nil, err
		}
	}

	pullReader, err := session.client.ImagePull(session.ctx, image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	_, err = ioutil.ReadAll(pullReader)
	if err != nil {
		return nil, err
	}
	err = pullReader.Close()
	if err != nil {
		return nil, err
	}

	config.Config.ContainerConfig.AttachStdin = true
	config.Config.ContainerConfig.AttachStderr = true
	config.Config.ContainerConfig.AttachStdout = true
	config.Config.ContainerConfig.OpenStdin = true
	config.Config.ContainerConfig.StdinOnce = true
	config.Config.ContainerConfig.Tty = session.pty
	if program != "" {
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
		return nil, err
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
		return nil, err
	}

	startOptions := types.ContainerStartOptions{}
	err = session.client.ContainerStart(session.ctx, session.containerId, startOptions)
	if err != nil {
		return nil, err
	}

	return &backend.ShellOrSubsystem{
		Stdin:  attachResult.Conn,
		Stdout: attachResult.Reader,
		//todo stderr handling?
		Stderr: attachResult.Reader,
	}, nil
}
