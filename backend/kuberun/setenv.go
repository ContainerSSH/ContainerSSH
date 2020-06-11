package kuberun

import "fmt"

func (session *kubeRunSession) SetEnv(name string, value string) error {
	if session.pod != nil {
		return fmt.Errorf("cannot set environment after the container has been started")
	}
	session.env[name] = value
	return nil
}
