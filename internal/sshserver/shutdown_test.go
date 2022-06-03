package sshserver_test

import (
	"context"
	"testing"
	"time"

    "go.containerssh.io/libcontainerssh/internal/sshserver"
    "go.containerssh.io/libcontainerssh/log"
)

func TestProperShutdown(t *testing.T) {
	user := sshserver.NewTestUser("foo")
	user.RandomPassword()
	logger := log.NewTestLogger(t)
	testServer := sshserver.NewTestServer(
		t,
		sshserver.NewTestAuthenticationHandler(
			sshserver.NewTestHandler(),
			user,
		),
		logger,
		nil,
	)
	testServer.Start()

	testClient := sshserver.NewTestClient(testServer.GetListen(), testServer.GetHostKey(), user, logger)
	connection := testClient.MustConnect()
	session := connection.MustSession()
	session.MustRequestPTY("xterm", 80, 25)
	session.MustShell()
	_ = session.WaitForStdout(context.Background(), []byte("> "))

	finished := make(chan struct{})
	go func() {
		testServer.Stop(60 * time.Second)
		finished <- struct{}{}
	}()
	go func() {
		_ = session.Wait()
		if err := connection.Close(); err != nil {
			t.Errorf("%v", err)
		}
	}()
	select {
	case <-time.After(30 * time.Second):
		t.Fail()
	case <-finished:
	}
}
