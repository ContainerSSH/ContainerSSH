package dockerrun

import "github.com/docker/docker/api/types"

func (session *dockerRunSession) removeContainer() error {
	if session.containerId != "" {
		//Update the exit code before destroying the container
		session.GetExitCode()
		removeOptions := types.ContainerRemoveOptions{Force: true}
		err := session.client.ContainerRemove(session.ctx, session.containerId, removeOptions)
		if err != nil {
			session.metric.Increment(MetricBackendError)
			return err
		}
	}
	return nil
}

func (session *dockerRunSession) Close() {
	_ = session.removeContainer()
	_ = session.client.Close()
}
