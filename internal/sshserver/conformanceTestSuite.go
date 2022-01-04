package sshserver

import (
	"context"
	"testing"
	"time"

	"github.com/containerssh/libcontainerssh/log"
)

type conformanceTestSuite struct {
	backendFactory ConformanceTestBackendFactory
}

func (c *conformanceTestSuite) singleProgramShouldRun(t *testing.T) {
	//t.Parallel()()
	logger := log.NewTestLogger(t)

	backend, err := c.backendFactory(t, logger)
	if err != nil {
		t.Fatal(err)
	}

	user := NewTestUser("test")
	user.RandomPassword()
	srv := NewTestServer(t, NewTestAuthenticationHandler(
		newConformanceTestHandler(backend),
		user,
	), logger, nil)
	srv.Start()
	defer srv.Stop(1 * time.Minute)

	client := NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	connection, err := client.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = connection.Close()
	}()
	session := connection.MustSession()
	if err := session.Exec("echo \"Hello world!\""); err != nil {
		t.Fatal(err)
	}
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := session.WaitForStdout(timeout, []byte("Hello world!\n")); err != nil {
		t.Fatal(err)
	}
	if err := session.Wait(); err != nil {
		t.Fatal(err)
	}
	if session.ExitCode() != 0 {
		t.Fatalf("invalid exit code returned: %d", session.ExitCode())
	}
	_ = session.Close()
	t.Log("test complete")
}

func (c *conformanceTestSuite) settingEnvVariablesShouldWork(t *testing.T) {
	//t.Parallel()()
	logger := log.NewTestLogger(t)
	backend, err := c.backendFactory(t, logger)
	if err != nil {
		t.Fatal(err)
	}

	user := NewTestUser("test")
	user.RandomPassword()
	srv := NewTestServer(t, NewTestAuthenticationHandler(
		newConformanceTestHandler(backend),
		user,
	), logger, nil)
	srv.Start()
	defer srv.Stop(1 * time.Minute)

	client := NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	connection, err := client.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = connection.Close()
	}()
	session := connection.MustSession()
	if err := session.SetEnv("FOO", "Hello world!"); err != nil {
		t.Fatal(err)
	}
	if err := session.Exec("echo \"$FOO\""); err != nil {
		t.Fatal(err)
	}
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := session.WaitForStdout(timeout, []byte("Hello world!\n")); err != nil {
		t.Fatal(err)
	}
	if err := session.Wait(); err != nil {
		t.Fatal(err)
	}
	if session.ExitCode() != 0 {
		t.Fatalf("invalid exit code returned: %d", session.ExitCode())
	}
	_ = session.Close()
	t.Log("test complete")
}

func (c *conformanceTestSuite) runningInteractiveShellShouldWork(t *testing.T) {
	//t.Parallel()()
	logger := log.NewTestLogger(t)
	backend, err := c.backendFactory(t, logger)
	if err != nil {
		t.Fatal(err)
	}

	user := NewTestUser("test")
	user.RandomPassword()
	srv := NewTestServer(t, NewTestAuthenticationHandler(
		newConformanceTestHandler(backend),
		user,
	), logger, nil)
	srv.Start()
	defer srv.Stop(1 * time.Minute)

	client := NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	connection, err := client.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = connection.Close()
	}()
	session := connection.MustSession()
	if err := session.SetEnv("foo", "bar"); err != nil {
		t.Error(err)
		return
	}
	if err := session.RequestPTY("xterm", 80, 25); err != nil {
		t.Error(err)
		return
	}
	if err := session.Shell(); err != nil {
		t.Error(err)
		return
	}
	if c.testShellInteraction(t, session) {
		return
	}
	if err := session.Wait(); err != nil {
		t.Error(err)
		return
	}
	if session.ExitCode() != 0 {
		t.Errorf("invalid exit code returned: %d", session.ExitCode())
		return
	}
	_ = session.Close()
	t.Log("test complete")
}

func (c *conformanceTestSuite) testShellInteraction(t *testing.T, session TestClientSession) bool {
	session.ReadRemaining()
	if !conformanceShellCommand(t, session, "tput cols", "80\r\n") {
		return true
	}
	if !conformanceShellCommand(t, session, "tput lines", "25\r\n") {
		return true
	}
	if err := session.Window(120, 25); err != nil {
		t.Error(err)
		return true
	}
	// Give Kubernetes time to realize the window change. Docker doesn't need this.
	time.Sleep(time.Second)
	// Read any output after the window change
	session.ReadRemaining()
	if !conformanceShellCommand(t, session, "tput cols", "120\r\n") {
		return true
	}
	if !conformanceShellCommand(t, session, "tput lines", "25\r\n") {
		return true
	}
	if !conformanceShellCommand(t, session, "echo \"Hello world!\"", "Hello world!\r\n") {
		return true
	}
	if !conformanceShellCommand(t, session, "exit", "") {
		return true
	}
	return false
}

func (c *conformanceTestSuite) reportingExitCodeShouldWork(t *testing.T) {
	//t.Parallel()()
	logger := log.NewTestLogger(t)
	backend, err := c.backendFactory(t, logger)
	if err != nil {
		t.Fatal(err)
	}

	user := NewTestUser("test")
	user.RandomPassword()
	srv := NewTestServer(t, NewTestAuthenticationHandler(
		newConformanceTestHandler(backend),
		user,
	), logger, nil)
	srv.Start()
	defer srv.Stop(1 * time.Minute)

	client := NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	connection, err := client.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = connection.Close()
	}()
	session := connection.MustSession()
	if err := session.Exec("exit 42"); err != nil {
		t.Fatal(err)
	}
	if err := session.Wait(); err != nil {
		t.Fatal(err)
	}
	if session.ExitCode() != 42 {
		t.Fatalf("invalid exit code returned: %d", session.ExitCode())
	}
	_ = session.Close()
	t.Log("test complete")
}

func (c *conformanceTestSuite) sendingSignalsShouldWork(t *testing.T) {
	//t.Parallel()()
	logger := log.NewTestLogger(t)
	backend, err := c.backendFactory(t, logger)
	if err != nil {
		t.Fatal(err)
	}

	user := NewTestUser("test")
	user.RandomPassword()
	srv := NewTestServer(t, NewTestAuthenticationHandler(
		newConformanceTestHandler(backend),
		user,
		nil,
	), logger, nil)
	srv.Start()
	defer srv.Stop(1 * time.Minute)

	client := NewTestClient(srv.GetListen(), srv.GetHostKey(), user, logger)
	connection, err := client.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = connection.Close()
	}()
	session := connection.MustSession()
	if err := session.Exec("/usr/bin/containerssh-agent wait-signal --signal USR1 --message \"USR1 received\""); err != nil {
		t.Fatal(err)
	}
	// Wait for backing program to start
	time.Sleep(time.Second)
	if err := session.Signal("USR1"); err != nil {
		t.Fatal(err)
	}
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := session.WaitForStdout(timeout, []byte("USR1 received")); err != nil {
		t.Fatal(err)
	}
	session.ReadRemaining()
	if err := session.Wait(); err != nil {
		t.Fatal(err)
	}
	if session.ExitCode() != 0 {
		t.Fatalf("invalid exit code returned: %d", session.ExitCode())
	}
	_ = session.Close()
	t.Log("test complete")
}
