package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

    config2 "go.containerssh.io/libcontainerssh/config"
    "go.containerssh.io/libcontainerssh/internal/metrics"
    "go.containerssh.io/libcontainerssh/log"
    "go.containerssh.io/libcontainerssh/message"
	core "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	watchTools "k8s.io/client-go/tools/watch"
)

type kubernetesPodImpl struct {
	config                config2.KubernetesConfig
	pod                   *core.Pod
	client                *kubernetes.Clientset
	restClient            *restclient.RESTClient
	logger                log.Logger
	tty                   *bool
	connectionConfig      *restclient.Config
	backendRequestsMetric metrics.SimpleCounter
	backendFailuresMetric metrics.SimpleCounter
	wg                    *sync.WaitGroup
	lock                  *sync.Mutex
	removeLock            *sync.Mutex
	shuttingDown          bool
	shutdown              bool
}

func (k *kubernetesPodImpl) getExitCode(ctx context.Context) (int32, error) {
	var pod *core.Pod
	var lastError error
loop:
	for {
		retryTimer := 10 * time.Second
		pod, lastError = k.client.CoreV1().Pods(k.pod.Namespace).Get(ctx, k.pod.Name, meta.GetOptions{})
		if lastError == nil {
			containerStatus := pod.Status.ContainerStatuses[k.config.Pod.ConsoleContainerNumber]
			if containerStatus.State.Terminated != nil {
				return containerStatus.State.Terminated.ExitCode, nil
			}
			lastError = fmt.Errorf("pod has not terminated yet")
			k.logger.Debug(
				message.Wrap(
					lastError,
					message.EKubernetesFetchingExitCodeFailed,
					"failed to fetch pod exit status, retrying in 10 seconds",
				))
			retryTimer = time.Second
		} else {
			if kubeErrors.IsNotFound(lastError) {
				err := message.Wrap(
					lastError,
					message.EKubernetesFetchingExitCodeFailed,
					"failed to fetch pod exit status, pod already removed",
				)
				k.logger.Debug(
					err,
				)
				return 137, err
			}
			k.logger.Debug(
				message.Wrap(
					lastError,
					message.EKubernetesFetchingExitCodeFailed,
					"failed to fetch pod exit status, retrying in 10 seconds",
				),
			)
		}
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(retryTimer):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := message.Wrap(
		lastError,
		message.EKubernetesFetchingExitCodeFailed,
		"failed to fetch pod exit status, giving up",
	)
	k.logger.Error(
		err,
	)
	return -1, err
}

func (k *kubernetesPodImpl) attach(_ context.Context) (kubernetesExecution, error) {
	k.logger.Debug(message.NewMessage(message.MKubernetesPodAttach, "attaching to pod..."))

	req := k.restClient.Post().
		Namespace(k.pod.Namespace).
		Resource("pods").
		Name(k.pod.Name).
		SubResource("attach")
	req.VersionedParams(
		&core.PodAttachOptions{
			Container: k.pod.Spec.Containers[k.config.Pod.ConsoleContainerNumber].Name,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       *k.tty,
		}, scheme.ParameterCodec,
	)

	podExec, err := remotecommand.NewSPDYExecutor(k.connectionConfig, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	return &kubernetesExecutionImpl{
		pod:  k,
		exec: podExec,
		terminalSizeQueue: &pushSizeQueueImpl{
			resizeChan: make(chan remotecommand.TerminalSize),
		},
		logger:                k.logger,
		tty:                   *k.tty,
		backendRequestsMetric: k.backendRequestsMetric,
		backendFailuresMetric: k.backendFailuresMetric,
		doneChan:              make(chan struct{}),
		lock:                  &sync.Mutex{},
	}, nil
}

func (k *kubernetesPodImpl) createExec(
	ctx context.Context,
	program []string,
	env map[string]string,
	tty bool,
) (kubernetesExecution, error) {
	k.lock.Lock()
	if k.shuttingDown {
		k.lock.Unlock()
		return nil, message.UserMessage(
			message.EKubernetesShuttingDown,
			"Server is shutting down",
			"Refusing new Kubernetes execution because the pod is shutting down.",
		)
	}
	k.wg.Add(1)
	k.lock.Unlock()
	exec, err := k.createExecLocked(ctx, program, env, tty)
	if err != nil {
		k.wg.Done()
	}
	return exec, err
}

func (k *kubernetesPodImpl) createExecLocked(
	_ context.Context,
	program []string,
	env map[string]string,
	tty bool,
) (kubernetesExecution, error) {
	k.logger.Debug(message.NewMessage(message.MKubernetesExec, "Creating and attaching to pod exec..."))

	if !k.config.Pod.DisableAgent {
		newProgram := []string{
			k.config.Pod.AgentPath,
			"console",
			"--pid",
		}
		for envKey, envValue := range env {
			newProgram = append(newProgram, "--env", fmt.Sprintf("%s=%s", envKey, envValue))
		}
		newProgram = append(newProgram, "--")
		program = append(newProgram, program...)
	}

	req := k.restClient.Post().
		Resource("pods").
		Name(k.pod.Name).
		Namespace(k.pod.Namespace).
		SubResource("exec")
	req.VersionedParams(
		&core.PodExecOptions{
			Container: k.pod.Spec.Containers[k.config.Pod.ConsoleContainerNumber].Name,
			Command:   program,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       tty,
		},
		scheme.ParameterCodec,
	)

	podExec, err := remotecommand.NewSPDYExecutor(
		k.connectionConfig,
		"POST",
		req.URL(),
	)
	if err != nil {
		return nil, err
	}

	return &kubernetesExecutionImpl{
		pod:  k,
		exec: podExec,
		terminalSizeQueue: &pushSizeQueueImpl{
			resizeChan: make(chan remotecommand.TerminalSize),
		},
		logger:                k.logger,
		env:                   env,
		tty:                   tty,
		backendRequestsMetric: k.backendRequestsMetric,
		backendFailuresMetric: k.backendFailuresMetric,
		doneChan:              make(chan struct{}),
		lock:                  &sync.Mutex{},
	}, nil
}

func (k *kubernetesPodImpl) writeFile(ctx context.Context, path string, content []byte) error {
	writeCmd := []string{k.config.Pod.AgentPath, "write-file", path}
	if k.config.Pod.Mode == config2.KubernetesExecutionModeSession {
		writeCmd = append(
			[]string{k.config.Pod.AgentPath, "console", "--wait", "--"},
			writeCmd...,
		)
	}
	exec, err := k.createExec(
		ctx,
		writeCmd,
		map[string]string{},
		false,
	)
	if err != nil {
		return err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdin := bytes.NewReader(content)

	exec.run(
		stdin,
		&stdout,
		&stderr,
		func() error {
			return nil
		},
		func(exitStatus int) {
		},
	)
	return nil
}

func (k *kubernetesPodImpl) remove(ctx context.Context) error {
	k.removeLock.Lock()
	defer k.removeLock.Unlock()
	if k.shuttingDown {
		return nil
	}

	k.lock.Lock()
	k.shuttingDown = true
	k.lock.Unlock()
	// Do not wait to exit in connection mode.
	// In session mode this function is called everytime an exec exits in order to ensure all signals/messages have been handled.
	// In connection mode this is only called in ssh disconnect, in which case we want to stop all activity
	if k.config.Pod.Mode == config2.KubernetesExecutionModeSession {
		k.wg.Wait()
	}
	k.lock.Lock()
	k.shutdown = true
	k.lock.Unlock()

	k.logger.Debug(message.NewMessage(message.MKubernetesPodRemove, "Removing pod..."))

	var lastError error
loop:
	for {
		lastError = k.client.CoreV1().Pods(k.pod.Namespace).Delete(ctx, k.pod.Name, meta.DeleteOptions{})
		if lastError == nil || kubeErrors.IsNotFound(lastError) {
			k.logger.Debug(message.NewMessage(message.MKubernetesPodRemoveSuccessful, "Pod removed."))
			return nil
		}
		k.logger.Debug(
			message.Wrap(
				lastError,
				message.EKubernetesFailedPodRemove,
				"Failed to remove pod, retrying in 10 seconds...",
			))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err := message.Wrap(lastError, message.EKubernetesFailedPodRemove, "Failed to remove pod, giving up.")
	k.logger.Error(
		err,
	)
	return err
}

func (k *kubernetesPodImpl) wait(ctx context.Context) (kubernetesPod, error) {
	k.logger.Debug(message.NewMessage(message.MKubernetesPodWait, "Waiting for pod to come up..."))

	k.backendRequestsMetric.Increment()
	fieldSelector := fields.
		OneTermEqualSelector("metadata.name", k.pod.Name).
		String()
	listWatch := &cache.ListWatch{
		ListFunc: func(options meta.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return k.client.
				CoreV1().
				Pods(k.pod.Namespace).
				List(ctx, options)
		},
		WatchFunc: func(options meta.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return k.client.
				CoreV1().
				Pods(k.pod.Namespace).
				Watch(ctx, options)
		},
	}

	event, err := watchTools.UntilWithSync(
		ctx,
		listWatch,
		&core.Pod{},
		nil,
		k.isPodAvailableEvent,
	)
	if event != nil {
		k.pod = event.Object.(*core.Pod)
	}
	if err != nil {
		err = message.WrapUser(
			err,
			message.MKubernetesPodWaitFailed,
			UserMessageInitializeSSHSession,
			"Failed to wait for pod to come up.",
		)
		k.logger.Error(err)
		k.backendFailuresMetric.Increment()
		return k, err
	}
	return k, err
}

func (k *kubernetesPodImpl) isPodAvailableEvent(event watch.Event) (bool, error) {
	if event.Type == watch.Deleted {
		return false, kubeErrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")
	}

	switch eventObject := event.Object.(type) {
	case *core.Pod:
		switch eventObject.Status.Phase {
		case core.PodFailed, core.PodSucceeded:
			return true, nil
		case core.PodRunning:
			conditions := eventObject.Status.Conditions
			for _, condition := range conditions {
				if condition.Type == core.PodReady &&
					condition.Status == core.ConditionTrue {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
