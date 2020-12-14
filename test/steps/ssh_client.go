package steps

import (
	"fmt"
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
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
	_ = conn.Close()
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
	_ = conn.Close()
	return nil
}

func (scenario *Scenario) RunCommand(username string, password string) error {
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
	defer func() {
		_ = conn.Close()
	}()

	sess, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = sess.Close()
	}()

	_, err = sess.StdinPipe()
	if err != nil {
		return err
	}

	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return err
	}

	_, err = sess.StderrPipe()
	if err != nil {
		return err
	}

	err = sess.Run("echo 'Hello world!'")
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(sessStdOut)
	if err != nil {
		return err
	}
	if string(bytes) != "Hello world!\n" {
		return fmt.Errorf("unexpected output (%s)", string(bytes))
	}

	return nil
}
