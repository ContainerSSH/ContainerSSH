package dockerrun

import "github.com/docker/docker/api/types"

func (session *dockerRunSession) removeContainer() error {
	if session.containerId != "" {
		removeOptions := types.ContainerRemoveOptions{Force: true}
		err := session.client.ContainerRemove(session.ctx, session.containerId, removeOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *dockerRunSession) Close() {
	_ = session.removeContainer()
	_ = session.client.Close()
}
