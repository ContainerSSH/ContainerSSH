package dockerrun

import (
)

func (session *dockerRunSession) GetExitCode() int32 {
	if session.exitCode < 0 && session.containerId != "" {
		inspect, err := session.client.ContainerInspect(session.ctx, session.containerId)
		if err != nil {
			session.logger.WarningF("error getting exit code from container (%v)", err)
		} else {
			session.exitCode = int32(inspect.State.ExitCode)
		}
	}
	return session.exitCode
}
