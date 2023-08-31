package kubernetes

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"go.containerssh.io/libcontainerssh/auth"
	publicConfig "go.containerssh.io/libcontainerssh/config"
	"go.containerssh.io/libcontainerssh/internal/agentforward"
	"go.containerssh.io/libcontainerssh/internal/sshserver"
	"go.containerssh.io/libcontainerssh/log"
	"go.containerssh.io/libcontainerssh/message"
	"go.containerssh.io/libcontainerssh/metadata"
)

type networkHandler struct {
	sshserver.AbstractNetworkConnectionHandler

	mutex        *sync.Mutex
	client       net.TCPAddr
	connectionID string
	config       publicConfig.KubernetesConfig

	cli          kubernetesClient
	pod          kubernetesPod
	logger       log.Logger
	disconnected bool
	labels       map[string]string
	annotations  map[string]string
	done         chan struct{}
}

func (n *networkHandler) OnAuthPassword(meta metadata.ConnectionAuthPendingMetadata, _ []byte) (
	sshserver.AuthResponse,
	metadata.ConnectionAuthenticatedMetadata,
	error,
) {
	return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf("the backend handler does not support authentication")
}

func (n *networkHandler) OnAuthPubKey(
	meta metadata.ConnectionAuthPendingMetadata,
	key auth.PublicKey,
) (sshserver.AuthResponse, metadata.ConnectionAuthenticatedMetadata, error) {
	return sshserver.AuthResponseUnavailable, meta.AuthFailed(), fmt.Errorf("the backend handler does not support authentication")
}

func (n *networkHandler) OnHandshakeFailed(_ metadata.ConnectionMetadata, _ error) {
}

func (n *networkHandler) OnHandshakeSuccess(meta metadata.ConnectionAuthenticatedMetadata) (
	connection sshserver.SSHConnectionHandler,
	returnMeta metadata.ConnectionAuthenticatedMetadata,
	failureReason error,
) {
	n.mutex.Lock()
	if n.pod != nil {
		n.mutex.Unlock()
		return nil, meta, fmt.Errorf("handshake already complete")
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), n.config.Timeouts.PodStart)
	defer func() {
		cancelFunc()
		n.mutex.Unlock()
	}()

	spec := n.config.Pod.Spec

	env := map[string]string{}
	for k, v := range meta.GetEnvironment() {
		env[k] = v.Value
	}
	for authMetadataName, labelName := range n.config.Pod.ExposeAuthMetadataAsEnv {
		if value, ok := meta.GetMetadata()[authMetadataName]; ok {
			env[labelName] = value.Value
		}
	}

	spec.Containers[n.config.Pod.ConsoleContainerNumber].Command = n.config.Pod.IdleCommand
	var r *regexp.Regexp
	var err error
	r, err = regexp.Compile(`[^[:alnum:]|\.]`)
	if err != nil {
		return nil, meta, fmt.Errorf("fail to compile regexp")
	}
	n.labels = map[string]string{
		"containerssh_connection_id": n.connectionID,
	}
	if len(meta.AuthenticatedUsername) <= 63 {
		n.labels["containerssh_username"] = r.ReplaceAllString(meta.AuthenticatedUsername, "-")
	} else {
		n.logger.Warning(message.NewMessage(
			message.MKubernetesUsernameTooLong,
			"The users username (%s) is longer than the 63 character limit of kubernetes labels. The containerssh_username label will be unavailable in the users pod",
			meta.AuthenticatedUsername,
		),
		)
	}
	for authMetadataName, labelName := range n.config.Pod.ExposeAuthMetadataAsLabels {
		if value, ok := meta.GetMetadata()[authMetadataName]; ok {
			n.labels[labelName] = value.Value
		}
	}

	n.annotations = map[string]string{
		"containerssh_ip": strings.ReplaceAll(n.client.IP.String(), ":", "-"),
	}
	for authMetadataName, annotationName := range n.config.Pod.ExposeAuthMetadataAsAnnotations {
		if value, ok := meta.GetMetadata()[authMetadataName]; ok {
			n.annotations[annotationName] = value.Value
		}
	}

	if n.config.Pod.Mode == publicConfig.KubernetesExecutionModeConnection {
		if n.pod, err = n.cli.createPod(ctx, n.labels, n.annotations, env, nil, nil); err != nil {
			return nil, meta, err
		}
		for path, content := range meta.GetFiles() {
			ctx, cancelFunc := context.WithTimeout(
				context.Background(),
				n.config.Timeouts.CommandStart,
			)
			n.logger.Debug(
				message.NewMessage(
					message.MKubernetesFileModification,
					"Writing to file %s",
					path,
				),
			)
			defer cancelFunc()
			err := n.pod.writeFile(ctx, path, content.Value)
			if err != nil {
				n.logger.Warning(
					message.Wrap(
						err,
						message.EKubernetesFileModificationFailed,
						"Failed to write to %s",
						path,
					),
				)
			}
		}
	}

	files := map[string][]byte{}
	for name, f := range meta.GetFiles() {
		files[name] = f.Value
	}

	return &sshConnectionHandler{
		networkHandler: n,
		username:       meta.Username,
		env:            env,
		files:          files,
		agentForward:   agentforward.NewAgentForward(n.logger),
	}, meta, nil
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
