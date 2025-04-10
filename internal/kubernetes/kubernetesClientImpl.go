package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	containerSSHConfig "go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/log"
	"go.containerssh.io/containerssh/message"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type kubernetesClientImpl struct {
	config                containerSSHConfig.KubernetesConfig
	logger                log.Logger
	client                *kubernetes.Clientset
	restClient            *restclient.RESTClient
	connectionConfig      *restclient.Config
	backendRequestsMetric metrics.SimpleCounter
	backendFailuresMetric metrics.SimpleCounter
}

func (k *kubernetesClientImpl) createPod(
	ctx context.Context,
	labels map[string]string,
	annotations map[string]string,
	env map[string]string,
	tty *bool,
	cmd []string,
) (kubePod kubernetesPod, lastError error) {
	podConfig, err := k.getPodConfig(tty, cmd, labels, annotations, env)
	if err != nil {
		return nil, err
	}
	logger := k.logger

	logger.Debug(message.NewMessage(message.MKubernetesPodCreate, "Creating pod"))
loop:
	for {
		kubePod, lastError = k.attemptPodCreate(ctx, podConfig, logger, tty)
		if lastError == nil {
			return kubePod, nil
		}
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	err = message.WrapUser(
		lastError,
		message.EKubernetesFailedPodCreate,
		UserMessageInitializeSSHSession,
		"Failed to create pod, giving up",
	)
	logger.Error(err)
	return nil, err
}

func (k *kubernetesClientImpl) getPodByName(
	ctx context.Context,
	podName string,
	namespace string,
) (*core.Pod, error) {

	return k.client.CoreV1().Pods(namespace).Get(ctx, podName, meta.GetOptions{})
}

func (k *kubernetesClientImpl) k8sPodToPodImpl(pod core.Pod) kubernetesPodImpl {
	tty := true
	return kubernetesPodImpl{
		pod:                   &pod,
		client:                k.client,
		restClient:            k.restClient,
		config:                k.config,
		logger:                k.logger.WithLabel("podName", pod.Name),
		connectionConfig:      k.connectionConfig,
		backendRequestsMetric: k.backendRequestsMetric,
		backendFailuresMetric: k.backendFailuresMetric,
		tty:                   &tty,
		lock:                  &sync.Mutex{},
		wg:                    &sync.WaitGroup{},
		removeLock:            &sync.Mutex{},
	}
}

func (k *kubernetesClientImpl) findPod(
	ctx context.Context,
	podName string,
	namespace string,
) (kubernetesPod, error) {
	for tries := 0; tries < 3; tries++ {
		pod, err := k.getPodByName(ctx, podName, namespace)
		if err != nil {
			k.logger.Debug("Error fetching pod: %v (retry %d)", err, tries+1)
			time.Sleep(2 * time.Second)
			continue
		}
		if pod == nil {
			k.logger.Debug(message.EKubernetesPodNotFound, "Pod not found, retrying... (retry %d)", tries+1)
			time.Sleep(2 * time.Second)
			continue
		}

		podImpl := k.k8sPodToPodImpl(*pod)
		return &podImpl, nil
	}
	notFoundErr := fmt.Errorf("failed to find pod %s in namespace %s after multiple retries",
		podName, namespace)
	k.logger.Debug(message.NewMessage(message.EKubernetesPodNotFound, notFoundErr.Error()))
	return nil, notFoundErr
}

func (k *kubernetesClientImpl) attemptPodCreate(
	ctx context.Context,
	podConfig containerSSHConfig.KubernetesPodConfig,
	logger log.Logger,
	tty *bool,
) (kubernetesPod, error) {
	var pod *core.Pod
	var lastError error
	k.backendRequestsMetric.Increment()
	pod, lastError = k.client.CoreV1().Pods(podConfig.Metadata.Namespace).Create(
		ctx,
		&core.Pod{
			ObjectMeta: podConfig.Metadata,
			Spec:       podConfig.Spec,
		},
		meta.CreateOptions{},
	)
	if lastError == nil {
		createdPod := &kubernetesPodImpl{
			pod:                   pod,
			client:                k.client,
			restClient:            k.restClient,
			config:                k.config,
			logger:                logger.WithLabel("podName", pod.Name),
			tty:                   tty,
			connectionConfig:      k.connectionConfig,
			backendRequestsMetric: k.backendRequestsMetric,
			backendFailuresMetric: k.backendFailuresMetric,
			lock:                  &sync.Mutex{},
			wg:                    &sync.WaitGroup{},
			removeLock:            &sync.Mutex{},
		}
		return createdPod.wait(ctx)
	}
	k.backendFailuresMetric.Increment()
	logger.Debug(
		message.Wrap(
			lastError,
			message.EKubernetesFailedPodCreate,
			"Failed to create pod, retrying in 10 seconds",
		),
	)
	return nil, lastError
}

func (k *kubernetesClientImpl) getPodConfig(
	tty *bool,
	cmd []string,
	labels map[string]string,
	annotations map[string]string,
	env map[string]string,
) (
	containerSSHConfig.KubernetesPodConfig,
	error,
) {
	var podConfig containerSSHConfig.KubernetesPodConfig
	if err := structutils.Copy(&podConfig, k.config.Pod); err != nil {
		return containerSSHConfig.KubernetesPodConfig{}, err
	}

	if podConfig.Mode == containerSSHConfig.KubernetesExecutionModeSession {
		if tty != nil {
			podConfig.Spec.Containers[k.config.Pod.ConsoleContainerNumber].TTY = *tty
		}
		podConfig.Spec.Containers[k.config.Pod.ConsoleContainerNumber].Stdin = true
		podConfig.Spec.Containers[k.config.Pod.ConsoleContainerNumber].StdinOnce = true
		if !podConfig.DisableAgent {
			podConfig.Spec.Containers[k.config.Pod.ConsoleContainerNumber].Command = append(
				[]string{
					podConfig.AgentPath,
					"console",
					"--wait",
					"--pid",
					"--",
				},
				cmd...,
			)
		} else {
			podConfig.Spec.Containers[k.config.Pod.ConsoleContainerNumber].Command = cmd
		}
		if podConfig.Spec.RestartPolicy == "" {
			podConfig.Spec.RestartPolicy = core.RestartPolicyNever
		}
	} else {
		podConfig.Spec.Containers[k.config.Pod.ConsoleContainerNumber].Command = k.config.Pod.IdleCommand
	}

	k.addLabelsToPodConfig(&podConfig, labels)
	k.addAnnotationsToPodConfig(&podConfig, annotations)
	k.addEnvToPodConfig(env, &podConfig)
	return podConfig, nil
}

func (k *kubernetesClientImpl) addLabelsToPodConfig(podConfig *containerSSHConfig.KubernetesPodConfig, labels map[string]string) {
	if podConfig.Metadata.Labels == nil {
		podConfig.Metadata.Labels = map[string]string{}
	}
	for k, v := range labels {
		podConfig.Metadata.Labels[k] = v
	}
}

func (k *kubernetesClientImpl) addAnnotationsToPodConfig(podConfig *containerSSHConfig.KubernetesPodConfig, annotations map[string]string) {
	if podConfig.Metadata.Annotations == nil {
		podConfig.Metadata.Annotations = map[string]string{}
	}
	for k, v := range annotations {
		podConfig.Metadata.Annotations[k] = v
	}
}

func (k *kubernetesClientImpl) addEnvToPodConfig(env map[string]string, podConfig *containerSSHConfig.KubernetesPodConfig) {
	for key, value := range env {
		for containerNo := range podConfig.Spec.Containers {
			podConfig.Spec.Containers[containerNo].Env = append(
				podConfig.Spec.Containers[containerNo].Env,
				core.EnvVar{
					Name:  key,
					Value: value,
				},
			)
		}
		for containerNo := range podConfig.Spec.InitContainers {
			podConfig.Spec.InitContainers[containerNo].Env = append(
				podConfig.Spec.InitContainers[containerNo].Env,
				core.EnvVar{
					Name:  key,
					Value: value,
				},
			)
		}
	}
}
