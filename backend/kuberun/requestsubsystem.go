package kuberun

import (
	"fmt"
	"io"
)

func (session *kubeRunSession) RequestSubsystem(subsystemName string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) error {
	if subsystemBinary, ok := session.config.Pod.Subsystems[subsystemName]; ok {
		return session.RequestProgram(subsystemBinary, stdIn, stdOut, stdErr, done)
	} else {
		return fmt.Errorf("unsupported subsystem: %s", subsystemName)
	}
}
