package test_test

import (
	"fmt"
	"net"
	"testing"

    "go.containerssh.io/libcontainerssh/internal/test"
	"golang.org/x/crypto/ssh"
)

func TestSSH(t *testing.T) {
	sshServer := test.SSH(t)

	config := &ssh.ClientConfig{
		User: sshServer.Username(),
		Auth: []ssh.AuthMethod{
			ssh.Password(sshServer.Password()),
		},
		HostKeyAlgorithms: sshServer.HostKeyAlgorithms(),
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if fingerprint := ssh.FingerprintSHA256(key); sshServer.FingerprintSHA256() != fingerprint {
				return fmt.Errorf("invalid SSH fingerprint: %s expected: %s", fingerprint, sshServer.FingerprintSHA256())
			}
			return nil
		}),
	}
	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshServer.Host(), sshServer.Port()), config)
	if err != nil {
		t.Fatalf("failed to connect to SSH server (%v)", err)
	}
	_ = sshConn.Close()
}
