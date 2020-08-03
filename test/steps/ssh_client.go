package steps

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
)

func (scenario *Scenario) AuthenticationShouldFail(username string, password string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	conn, err := ssh.Dial("tcp", "127.0.0.1:2222", config)
	if err != nil {
		return nil
	}
	defer conn.Close()
	return fmt.Errorf("SSH connection did not fail")
}

func (scenario *Scenario) AuthenticationShouldSucceed(username string, password string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", "127.0.0.1:2222", config)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
