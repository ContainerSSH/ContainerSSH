package dockerrun

import (
	"containerssh/backend"
	"fmt"
)

func (session *dockerRunSession) RequestSubsystem(subsystem string) (*backend.ShellOrSubsystem, error) {
	return nil, fmt.Errorf("not implemented")
}
