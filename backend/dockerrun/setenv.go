package dockerrun

import "fmt"

func (session *dockerRunSession) SetEnv(name string, value string) error {
	if session.containerId != "" {
		return fmt.Errorf("cannot set environment after the container has been started")
	}
	session.env[name] = value
	return nil
}
