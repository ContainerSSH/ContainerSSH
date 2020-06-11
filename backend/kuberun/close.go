package kuberun

import (
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (session *kubeRunSession) removePod() error {
	if session.pod != nil {
		//TODO when the SSH session is being closed no exit code is transmitted because the container is still running.
		//Update the exit code before destroying the container
		session.GetExitCode()
		request := session.restClient.
			Delete().
			Namespace(session.pod.Namespace).
			Resource("pods").
			Name(session.pod.Name).
			Body(&metav1.DeleteOptions{})
		logrus.Tracef("Deleting %s", request.URL())
		result := request.Do(session.ctx)
		if result.Error() != nil {
			logrus.Tracef("Failed to remove pod (%s)", result.Error())
			return result.Error()
		}
		session.pod = nil
	}
	return nil
}

func (session *kubeRunSession) Close() {
	_ = session.removePod()
}
