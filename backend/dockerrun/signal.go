package dockerrun

import "fmt"

func (session *dockerRunSession) SendSignal(signal string) error {
	if session.containerId == "" {
		return fmt.Errorf("cannot send signal if a container is not running")
	}
	err := session.client.ContainerKill(session.ctx, session.containerId, signal)
	if err != nil {
		session.metric.Increment(MetricBackendError)
	}
	return err
}
