package dockerrun

import "fmt"

func (session *dockerRunSession) SetPty() error {
	if session.containerId != "" {
		return fmt.Errorf("cannot change pty mode after the container has started")
	}
	session.pty = true
	return nil
}
