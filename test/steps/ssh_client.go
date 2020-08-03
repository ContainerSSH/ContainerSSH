package steps

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

func (scenario *Scenario) AuthenticationShouldFail(username string, password string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	conn, err := ssh.Dial("tcp", "127.0.0.1:2222", config)
	if err != nil {
		return nil
	}
	defer conn.Close()
	return fmt.Errorf("SSH connection did not fail")
}