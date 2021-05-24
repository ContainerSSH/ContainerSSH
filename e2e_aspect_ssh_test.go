package containerssh_test

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/containerssh/configuration/v2"
	"github.com/containerssh/log"
	"golang.org/x/crypto/ssh"

	"github.com/containerssh/containerssh"
)

func NewSSHTestingAspect() TestingAspect {
	return &sshTestingAspect{}
}

type sshTestingAspect struct {
}

func (r *sshTestingAspect) String() string {
	return "SSH"
}

func (r *sshTestingAspect) Factors() []TestingFactor {
	var factor TestingFactor = &sshInProcess{
		aspect: r,
	}
	return []TestingFactor{
		factor,
	}
}

type sshInProcess struct {
	lifecycle     *SimpleLifecycle
	config        configuration.AppConfig
	sshConnection *ssh.Client
	session       *ssh.Session
	result        []byte
	exitCode      int
	aspect        *sshTestingAspect
}

func (r *sshInProcess) Aspect() TestingAspect {
	return r.aspect
}

func (r *sshInProcess) String() string {
	return "In-Process"
}

func (r *sshInProcess) ModifyConfiguration(*configuration.AppConfig) error {
	return nil
}

func (r *sshInProcess) StartBackingServices(
	config configuration.AppConfig,
	_ log.Logger,
) error {
	if err := config.SSH.GenerateHostKey(); err != nil {
		return err
	}
	r.config = config
	srv, err := containerssh.New(
		config,
		log.NewLoggerFactory(),
	)
	if err != nil {
		return err
	}
	r.lifecycle = NewSimpleLifecycle(srv)
	return r.lifecycle.Start()
}

func (r *sshInProcess) getConfig(user string, password string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
}

func (r *sshInProcess) StopBackingServices(_ configuration.AppConfig, _ log.Logger) error {
	if r.sshConnection != nil {
		_ = r.sshConnection.Close()
	}
	return r.lifecycle.Stop()
}

func (r *sshInProcess) GetSteps(_ configuration.AppConfig, _ log.Logger) []Step {
	return []Step{
		{
			`^authentication with user "(.*)" and password "(.*)" (?:should fail|should have failed)$`,
			r.AuthenticationShouldFail,
		},
		{
			`^authentication with user "(.*)" and password "(.*)" (?:should succeed|should have succeeded)$`,
			r.AuthenticationShouldSucceed,
		},
		{
			`^I open an SSH connection with the user "([^"]*)" and the password "([^"]*)"$`,
			r.OpenSSHConnection,
		},
		{
			`^I open an SSH session$`,
			r.OpenSSHSession,
		},
		{
			`^I set the environment variable "([^"]*)" to the value "([^"]*)"$`,
			r.SetEnvVariable,
		},
		{
			`^I execute the command "(.*)"$`,
			r.ExecuteTheCommand,
		},
		{
			`^I should see "([^"]*)" in the output$`,
			r.ShouldSeeOutput,
		},
		{
			`^the session should exit with the code "([^"]*)"$`,
			r.SessionShouldExitWithCode,
		},
	}
}

func (r *sshInProcess) AuthenticationShouldFail(user string, password string) error {
	sshConnection, err := ssh.Dial("tcp", r.config.SSH.Listen, r.getConfig(user, password))
	if err != nil {
		return nil
	}
	_ = sshConnection.Close()
	return fmt.Errorf("the authentication did not result in an error")
}

func (r *sshInProcess) AuthenticationShouldSucceed(user string, password string) error {
	sshConnection, err := ssh.Dial("tcp", r.config.SSH.Listen, r.getConfig(user, password))
	if err != nil {
		return fmt.Errorf("the authentication resulted in an error (%w)", err)
	}
	return sshConnection.Close()
}

func (r *sshInProcess) OpenSSHConnection(user string, password string) error {
	var err error
	r.sshConnection, err = ssh.Dial("tcp", r.config.SSH.Listen, r.getConfig(user, password))
	if err != nil {
		return fmt.Errorf("the authentication resulted in an error (%w)", err)
	}
	return nil
}

func (r *sshInProcess) OpenSSHSession() error {
	if r.sshConnection == nil {
		return fmt.Errorf("no SSH connection open")
	}
	if r.session != nil {
		_ = r.session.Close()
	}
	var err error
	r.session, err = r.sshConnection.NewSession()
	r.exitCode = -1
	if err != nil {
		return fmt.Errorf("cannot open session (%w)", err)
	}
	return nil
}

func (r *sshInProcess) SetEnvVariable(name string, value string) error {
	if r.session == nil {
		return fmt.Errorf("no SSH session open")
	}
	return r.session.Setenv(name, value)
}

func (r *sshInProcess) ExecuteTheCommand(command string) error {
	var err error
	r.result, err = r.session.CombinedOutput(command)
	exitError := &ssh.ExitError{}
	if err != nil {
		if errors.As(err, &exitError) {
			r.exitCode = exitError.ExitStatus()
			return nil
		}
		return err
	} else {
		r.exitCode = 0
		return nil
	}
}

func (r *sshInProcess) ShouldSeeOutput(expected string) error {
	if !strings.Contains(string(r.result), expected) {
		return fmt.Errorf("expected output not found in execution result:\n%s", r.result)
	}
	return nil
}

func (r *sshInProcess) SessionShouldExitWithCode(exitCode int) error {
	if r.exitCode != exitCode {
		return fmt.Errorf("unexpected exit code: %d", r.exitCode)
	}
	return nil
}
