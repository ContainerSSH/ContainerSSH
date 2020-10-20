package dockerrun

import "github.com/docker/docker/api/types"

func (session *dockerRunSession) Resize(cols uint, rows uint) error {
	session.cols = cols
	session.rows = rows
	if session.containerId != "" {
		resizeOptions := types.ResizeOptions{}
		resizeOptions.Width = session.cols
		resizeOptions.Height = session.rows
		err := session.client.ContainerResize(session.ctx, session.containerId, resizeOptions)
		if err != nil {
			session.metric.Increment(MetricBackendError)
			return err
		}
	}
	return nil
}
