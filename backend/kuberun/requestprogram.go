package kuberun

import (
	"context"
	"fmt"
	"github.com/janoszen/containerssh/config"
	"github.com/mattn/go-shellwords"
	"github.com/sirupsen/logrus"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/kubectl/pkg/util/interrupt"
	"strings"
)

//This function returns true if the pod is either running or already has finished running (in which case logs are
//available).
func (session *kubeRunSession) isPodAvailableEvent(event watch.Event) (bool, error) {
	if event.Type == watch.Deleted {
		return false, errors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")
	}

	switch eventObject := event.Object.(type) {
	case *v1.Pod:
		switch eventObject.Status.Phase {
		case v1.PodFailed, v1.PodSucceeded:
			return true, nil
		case v1.PodRunning:
			conditions := eventObject.Status.Conditions
			if conditions != nil {
				for _, condition := range conditions {
					if condition.Type == v1.PodReady &&
						condition.Status == v1.ConditionTrue {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

//This function waits for a pod to be either running or already complete.
func (session *kubeRunSession) waitForPodAvailable() (result *v1.Pod, err error) {
	timeoutContext, cancelTimeoutContext := watchtools.ContextWithOptionalTimeout(session.ctx, session.config.Timeout)
	defer cancelTimeoutContext()

	//TODO if metadata is made customizable this must be adjusted to match
	fieldSelector := fields.
		OneTermEqualSelector("metadata.name", session.pod.Name).
		String()
	listWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return session.client.
				CoreV1().
				Pods(session.pod.Namespace).
				List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return session.client.
				CoreV1().
				Pods(session.pod.Namespace).
				Watch(context.TODO(), options)
		},
	}

	err = interrupt.
		New(nil, cancelTimeoutContext).
		Run(
			func() error {
				event, err := watchtools.UntilWithSync(
					timeoutContext,
					listWatch,
					&v1.Pod{},
					nil,
					session.isPodAvailableEvent,
				)
				if event != nil {
					result = event.Object.(*v1.Pod)
				}
				return err
			},
		)

	return result, err
}

func (session *kubeRunSession) RequestProgram(program string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) error {
	if session.pod != nil {
		return fmt.Errorf("cannot change request program after the pod has started")
	}

	if session.config.Pod.DisableCommand && program != "" {
		return fmt.Errorf("command execution disabled, cannot run program: %s", program)
	}

	spec := session.createPodSpec(program, session.config)

	logrus.Tracef("Creating pod")
	pod, err := session.client.CoreV1().Pods(session.config.Pod.Namespace).Create(
		session.ctx,
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "containerssh-",
				Namespace:    session.config.Pod.Namespace,
			},
			Spec: spec,
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return err
	}

	session.pod = pod
	container := pod.Spec.Containers[session.config.Pod.ConsoleContainerNumber]

	go session.handlePodConnection(err, pod, container, stdIn, stdOut, stdErr, done)
	return nil
}

func (session *kubeRunSession) createPodSpec(program string, config config.KubeRunConfig) v1.PodSpec {
	spec := config.Pod.Spec

	consoleContainerNumber := config.Pod.ConsoleContainerNumber
	for key, value := range session.env {
		spec.Containers[consoleContainerNumber].Env = append(
			spec.Containers[consoleContainerNumber].Env,
			v1.EnvVar{
				Name:  key,
				Value: value,
			},
		)
	}

	spec.Containers[consoleContainerNumber].TTY = session.pty
	spec.Containers[consoleContainerNumber].StdinOnce = true
	spec.Containers[consoleContainerNumber].Stdin = true
	spec.RestartPolicy = v1.RestartPolicyNever

	if !config.Pod.DisableCommand && program != "" {
		programParts, err := shellwords.Parse(program)
		if err != nil {
			spec.Containers[consoleContainerNumber].Command = []string{"/bin/sh", "-c", program}
		} else {
			if strings.HasPrefix(programParts[0], "/") ||
				strings.HasPrefix(programParts[0], "./") ||
				strings.HasPrefix(programParts[0], "../") {
				spec.Containers[consoleContainerNumber].Command = programParts
			} else {
				spec.Containers[consoleContainerNumber].Command = []string{"/bin/sh", "-c", program}
			}
		}
	}
	return spec
}

func (session *kubeRunSession) handlePodConnection(err error, pod *v1.Pod, container v1.Container, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) {
	// This function is running async so that the SSH client can receive an acknowledgement of the requested
	// shell/exec/subsystem before sending output, otherwise the output will be ignored.
	// Todo possible race condition? Unlikely as the pod takes time to come up

	//Wait for pod do come up
	session.pod, err = session.waitForPodAvailable()
	if err != nil {
		logrus.Tracef("Pod failed to come up (%s)", err)
		session.Close()
		_ = session.removePod()
		//todo send error to client
	}

	req := session.restClient.Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("attach")
	req.VersionedParams(&v1.PodAttachOptions{
		Container: container.Name,
		Stdin:     true,
		Stdout:    true,
		Stderr:    !session.pty,
		TTY:       session.pty,
	}, scheme.ParameterCodec)

	logrus.Tracef("POST %s", req.URL())
	exec, err := remotecommand.NewSPDYExecutor(&session.connectionConfig, "POST", req.URL())
	if err != nil {
		logrus.Warnf("Failed to attach to container (%s)", err)
		session.Close()
		_ = session.removePod()
		return
	}

	logrus.Tracef("Streaming input/output")
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdIn,
		Stdout:            stdOut,
		Stderr:            stdErr,
		Tty:               session.pty,
		TerminalSizeQueue: &session.terminalSizeQueue,
	})
	if err != nil {
		session.handleFinishedPod(pod, container, stdOut)
	}
	session.Close()
	done()
}

func (session *kubeRunSession) handleFinishedPod(pod *v1.Pod, container v1.Container, stdOut io.Writer) {
	logrus.Tracef("Pod already finished, streaming logs")
	//Try fetching logs in case the container already exited.
	//Do not move this above the "attach" method otherwise there will be a race condition.
	request := session.client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container.Name,
	})

	logStream, err := request.Stream(session.ctx)
	if err != nil {
		logrus.Tracef("Failed to attach or stream logs (%s)", err)
		return
	}
	//TODO stderr stream? Does Kubernetes even support that with logs?
	//     https://github.com/kubernetes/kubernetes/issues/28167
	_, err = io.Copy(stdOut, logStream)
	if err != nil {
		logrus.Tracef("Failed to attach or stream logs (%s)", err)
		return
	}
	//todo inform client of the error.
}
