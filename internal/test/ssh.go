package test

import (
	"embed"
	"fmt"
	"testing"

	"golang.org/x/crypto/ssh"
)

//go:embed ssh
var sshBuildRoot embed.FS

// SSHHelper contains the details about a running SSH service.
type SSHHelper interface {
	// Host returns the IP address of the running SSH service.
	Host() string
	// Port returns the port number of the running SSH service.
	Port() int
	// FingerprintSHA256 returns the SHA 256 fingerprint of the server.
	FingerprintSHA256() string
	// HostKey returns the private SSH host key.
	HostKey() []byte
	// Username returns the SSH username to use for connecting.
	Username() string
	// Password returns the SSH password to use for connecting.
	Password() string
	// HostKeyAlgorithms returns a list of usable host key algorithms.
	HostKeyAlgorithms() []string
}

// SSH creates a fully functional SSH service which you can SSH into for testing purposes.
func SSH(t *testing.T) SSHHelper {
	username := "ubuntu"
	password := "ubuntu"
	files := dockerBuildRootFiles(sshBuildRoot, "ssh")
	cnt := containerFromBuild(
		t, "ssh", files, nil, []string{
			fmt.Sprintf("SSH_USERNAME=%s", username),
			fmt.Sprintf("SSH_PASSWORD=%s", password),
		},
		map[string]string{
			"22/tcp": "",
		},
	)

	hostKey := cnt.extractFile("/etc/ssh/ssh_host_rsa_key")
	signer, err := ssh.ParsePrivateKey(hostKey)
	if err != nil {
		t.Fatalf("failed to parse SSH host key (%v)", err)
	}

	return &sshHelper{
		username:    username,
		password:    password,
		hostKey:     hostKey,
		signer:      signer,
		fingerprint: ssh.FingerprintSHA256(signer.PublicKey()),
		cnt:         cnt,
	}
}

type sshHelper struct {
	username    string
	password    string
	cnt         container
	signer      ssh.Signer
	fingerprint string
	hostKey     []byte
}

func (s *sshHelper) HostKeyAlgorithms() []string {
	return []string{
		s.signer.PublicKey().Type(),
	}
}

func (s *sshHelper) Username() string {
	return s.username
}

func (s *sshHelper) Password() string {
	return s.password
}

func (s *sshHelper) FingerprintSHA256() string {
	return s.fingerprint
}

func (s *sshHelper) HostKey() []byte {
	return s.hostKey
}

func (s *sshHelper) Host() string {
	return "127.0.0.1"
}

func (s *sshHelper) Port() int {
	return s.cnt.port("22/tcp")
}
