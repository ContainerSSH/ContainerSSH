package kubernetes

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/containerssh/libcontainerssh/auth"
	"github.com/containerssh/libcontainerssh/message"
	config2 "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/sshserver"
	"github.com/containerssh/libcontainerssh/log"
)

type networkHandler struct {
	sshserver.AbstractNetworkConnectionHandler

	mutex        *sync.Mutex
	client       net.TCPAddr
	connectionID string
	config       config2.KubernetesConfig

	cli          kubernetesClient
	pod          kubernetesPod
	logger       log.Logger
	disconnected bool
	labels       map[string]string
	annotations  map[string]string
	done         chan struct{}
}

func (n *networkHandler) OnAuthPassword(_ string, _ []byte, _ string) (response sshserver.AuthResponse, metadata *auth.ConnectionMetadata, reason error) {
	return sshserver.AuthResponseUnavailable, nil, fmt.Errorf("the backend handler does not support authentication")
}

func (n *networkHandler) OnAuthPubKey(_ string, _ string, _ string) (response sshserver.AuthResponse, metadata *auth.ConnectionMetadata, reason error) {
	return sshserver.AuthResponseUnavailable, nil, fmt.Errorf("the backend handler does not support authentication")
}

func (n *networkHandler) OnHandshakeFailed(_ error) {
}

func (n *networkHandler) OnHandshakeSuccess(username string, clientVersion string, metadata *auth.ConnectionMetadata) (connection sshserver.SSHConnectionHandler, failureReason error) {
	n.mutex.Lock()
	if n.pod != nil {
		n.mutex.Unlock()
		return nil, fmt.Errorf("handshake already complete")
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), n.config.Timeouts.PodStart)
	defer func() {
		cancelFunc()
		n.mutex.Unlock()
	}()

	spec := n.config.Pod.Spec

	env := map[string]string{}
	for authMetadataName, envName := range n.config.Pod.ExposeAuthMetadataAsEnv {
		if value, ok := metadata.GetMetadata()[authMetadataName]; ok {
			env[envName] = value
		}
	}
	for k, v := range metadata.GetEnvironment() {
		env[k] = v
	}

	spec.Containers[n.config.Pod.ConsoleContainerNumber].Command = n.config.Pod.IdleCommand
	n.labels = map[string]string{
		"containerssh_connection_id": n.connectionID,
		"containerssh_username":      username,
	}
	for authMetadataName, labelName := range n.config.Pod.ExposeAuthMetadataAsLabels {
		if value, ok := metadata.GetMetadata()[authMetadataName]; ok {
			n.labels[labelName] = value
		}
	}

	n.annotations = map[string]string{
		"containerssh_ip": strings.ReplaceAll(n.client.IP.String(), ":", "-"),
	}
	for authMetadataName, annotationName := range n.config.Pod.ExposeAuthMetadataAsAnnotations {
		if value, ok := metadata.GetMetadata()[authMetadataName]; ok {
			n.annotations[annotationName] = value
		}
	}

	var err error
	if n.config.Pod.Mode == config2.KubernetesExecutionModeConnection {
		if n.pod, err = n.cli.createPod(ctx, n.labels, n.annotations, env, nil, nil); err != nil {
			return nil, err
		}
		for path, content := range metadata.GetFiles() {
			ctx, cancelFunc := context.WithTimeout(
				context.Background(),
				n.config.Timeouts.CommandStart,
			)
			n.logger.Debug(message.NewMessage(
				message.MKubernetesFileModification,
				"Writing to file %s",
				path,
			))
			defer cancelFunc()
			err := n.pod.writeFile(ctx, path, content)
			if err != nil {
				n.logger.Warning(message.Wrap(
					err,
					message.EKubernetesFileModificationFailed,
					"Failed to write to %s",
					path,
				))
			}
		}
	}

	return &sshConnectionHandler{
		networkHandler: n,
		username:       username,
		env:            env,
		files:          metadata.GetFiles(),
	}, nil
}

func (n *networkHandler) OnDisconnect() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.disconnected {
		return
	}
	n.disconnected = true
	ctx, cancelFunc := context.WithTimeout(context.Background(), n.config.Timeouts.PodStop)
	defer cancelFunc()
	if n.pod != nil {
		_ = n.pod.remove(ctx)
	}
	close(n.done)
}

func (n *networkHandler) OnShutdown(shutdownContext context.Context) {
	select {
	case <-shutdownContext.Done():
		n.OnDisconnect()
	case <-n.done:
	}
}
