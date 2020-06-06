package dockerrun

import (
	"containerssh/backend"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func (session *dockerRunSession) RequestProgram(program string) (*backend.ShellOrSubsystem, error) {
	if session.containerId != "" {
		return nil, fmt.Errorf("cannot change request shell after the container has started")
	}

	//todo replace with configuration
	containerConfig := &container.Config{}
	containerConfig.Image = "busybox"
	containerConfig.Hostname = "test"
	containerConfig.Domainname = "example.com"
	containerConfig.AttachStdin = true
	containerConfig.AttachStderr = true
	containerConfig.AttachStdout = true
	containerConfig.OpenStdin = true
	containerConfig.StdinOnce = true
	containerConfig.Tty = session.pty
	containerConfig.NetworkDisabled = false
	if program != "" {
		containerConfig.Cmd = []string{"/bin/sh", "-c", program}
	}
	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{}
	body, err := session.client.ContainerCreate(session.ctx, containerConfig, hostConfig, networkingConfig, "")
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
