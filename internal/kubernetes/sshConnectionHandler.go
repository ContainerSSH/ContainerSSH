package kubernetes

import (
	"context"
	"fmt"
	"io"

	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/agentforward"
	"go.containerssh.io/containerssh/internal/sshserver"
	"go.containerssh.io/containerssh/message"
	"go.containerssh.io/containerssh/metadata"
	"golang.org/x/crypto/ssh"
)

type sshConnectionHandler struct {
	networkHandler *networkHandler
	username       string
	env            map[string]string
	files          map[string][]byte
	agentForward   agentforward.AgentForward
}

func (s *sshConnectionHandler) OnUnsupportedGlobalRequest(_ uint64, _ string, _ []byte) {
}

func (b *sshConnectionHandler) OnFailedDecodeGlobalRequest(_ uint64, _ string, _ []byte, _ error) {}

func (s *sshConnectionHandler) OnUnsupportedChannel(_ uint64, _ string, _ []byte) {
}

func (s *sshConnectionHandler) OnShutdown(_ context.Context) {}

func (s *sshConnectionHandler) OnSessionChannel(
	meta metadata.ChannelMetadata,
	_ []byte,
	session sshserver.SessionChannel,
) (
	channel sshserver.SessionChannelHandler,
	failureReason sshserver.ChannelRejection,
) {
	env := map[string]string{}
	for k, v := range s.env {
		env[k] = v
	}
	return &channelHandler{
		session:           session,
		channelID:         meta.ChannelID,
		networkHandler:    s.networkHandler,
		connectionHandler: s,
		username:          s.username,
		env:               env,
		files:             s.files,
	}, nil
}

func (s *sshConnectionHandler) OnTCPForwardChannel(
	channelID uint64,
	hostToConnect string,
	portToConnect uint32,
	originatorHost string,
	originatorPort uint32,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	channel, err := s.agentForward.NewForwardTCP(
		s.setupAgent,
		s.networkHandler.logger,
		hostToConnect,
		portToConnect,
		originatorHost,
		originatorPort,
	)
	if err != nil {
		return nil, sshserver.NewChannelRejection(ssh.ConnectionFailed, message.EKubernetesForwardingFailed, "Error setting up the forwarding", "Error setting up the forwarding")
	}
	return channel, nil
}

func (s *sshConnectionHandler) OnAuthAgentChannel(channelID uint64) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	channel, err := s.agentForward.NewForwardUnix(
		s.setupAgent,
		s.networkHandler.logger,
		"/tmp/ssh-agent.sock",
	)
	if err != nil {
		return nil, sshserver.NewChannelRejection(ssh.ConnectionFailed, message.EKubernetesForwardingFailed, "Error setting up SSH agent forwarding", "Error setting up SSH agent forwarding")
	}
	return channel, nil
}

func (s *sshConnectionHandler) OnRequestTCPReverseForward(
	bindHost string,
	bindPort uint32,
	reverseHandler sshserver.ReverseForward,
) error {
	return s.agentForward.NewTCPReverseForwarding(
		s.setupAgent,
		s.networkHandler.logger,
		bindHost,
		bindPort,
		reverseHandler,
	)
}

func (s *sshConnectionHandler) OnRequestCancelTCPReverseForward(
	bindHost string,
	bindPort uint32,
) error {
	return s.agentForward.CancelTCPForwarding(bindHost, bindPort)
}

func (s *sshConnectionHandler) OnDirectStreamLocal(
	channelID uint64,
	path string,
) (channel sshserver.ForwardChannel, failureReason sshserver.ChannelRejection) {
	channel, err := s.agentForward.NewForwardUnix(
		s.setupAgent,
		s.networkHandler.logger,
		path,
	)
	if err != nil {
		return nil, sshserver.NewChannelRejection(ssh.ConnectionFailed, message.EKubernetesForwardingFailed, "Error setting up the forwarding", "Error setting up the forwarding")
	}
	return channel, nil
}

func (s *sshConnectionHandler) OnRequestStreamLocal(
	path string,
	reverseHandler sshserver.ReverseForward,
) error {
	return s.agentForward.NewUnixReverseForwarding(
		s.setupAgent,
		s.networkHandler.logger,
		path,
		reverseHandler,
	)
}

func (s *sshConnectionHandler) OnRequestCancelStreamLocal(
	path string,
) error {
	return s.agentForward.CancelStreamLocalForwarding(path)
}

func (c *sshConnectionHandler) setupAgent() (io.Reader, io.Writer, error) {
	ctx, cancelFunc := context.WithTimeout(
		context.Background(),
		c.networkHandler.config.Timeouts.CommandStart,
	)
	defer cancelFunc()

	if c.networkHandler.config.Pod.Mode == config.KubernetesExecutionModeConnection {
		agent := []string{c.networkHandler.config.Pod.AgentPath, "forward-server"}
		exec, err := c.networkHandler.pod.createExec(ctx, agent, c.env, false)
		if err != nil {
			c.networkHandler.logger.Warning(message.Wrap(
				err,
				message.EKubernetesAgentFailed,
				"Failed to start the agent",
			))
			return nil, nil, err
		}

		stdinReader, stdinWriter := io.Pipe()
		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()

		go func() {
			buf := make([]byte, 8192)
			for {
				nBytes, err := stderrReader.Read(buf)
				if nBytes != 0 {
					c.networkHandler.logger.Info(
						message.NewMessage(
							message.MKubernetesAgentLog,
							"%s",
							string(buf[:nBytes]),
						),
					)
				}
				if err != nil {
					return
				}
			}
		}()

		exec.run(
			stdinReader,
			stdoutWriter,
			stderrWriter,
			func() error {
				_ = stdinWriter.Close()
				_ = stdoutReader.Close()
				_ = stderrReader.Close()
				return nil
			},
			func(exitStatus int) {
				if exitStatus != 0 {
					c.networkHandler.logger.Warning(message.NewMessage(
						message.EKubernetesAgentFailed,
						"Agent exited with exit code %d",
						exitStatus,
					))
				}
			},
		)
		return stdoutReader, stdinWriter, nil
	}

	return nil, nil, fmt.Errorf("port forwarding does not work in session mode")
}
