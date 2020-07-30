package dockerrun

import (
	"fmt"
	"io"
)

func (session *dockerRunSession) RequestSubsystem(subsystemName string, stdIn io.Reader, stdOut io.Writer, stdErr io.Writer, done func()) error {
	if session.config.Config.DisableCommand {
		return fmt.Errorf("subsystem disabled: %s", subsystemName)
	}
	if subsystemBinary, ok := session.config.Config.Subsystems[subsystemName]; ok {
		return session.RequestProgram(subsystemBinary, stdIn, stdOut, stdErr, done)
	} else {
		return fmt.Errorf("unsupported subsystem: %s", subsystemName)
	}
}
