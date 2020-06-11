package dockerrun

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

func (session *dockerRunSession) GetExitCode() int32 {
	if session.exitCode < 0 && session.containerId != "" {
		inspect, err := session.client.ContainerInspect(session.ctx, session.containerId)
		if err != nil {
			logrus.Warn(fmt.Sprintf("Error getting exit code from container (%s)", err))
		} else {
			session.exitCode = int32(inspect.State.ExitCode)
		}
	}
	return session.exitCode
}
