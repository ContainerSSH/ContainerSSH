package kuberun

import "fmt"

func (session *kubeRunSession) SetPty() error {
	if session.pod != nil {
		return fmt.Errorf("cannot change pty mode after the container has started")
	}
	session.pty = true
	return nil
}
