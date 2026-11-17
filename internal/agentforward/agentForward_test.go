package agentforward_test

import (
	"io"
	"testing"
	"time"

	protocol "go.containerssh.io/containerssh/agentprotocol"
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/agentforward"
	"go.containerssh.io/containerssh/log"
)

func TestOnShutdownX11Only(t *testing.T) {
	logger := discardLogger(t)
	forward := agentforward.NewAgentForward(logger)
	if err := forward.NewX11Forwarding(fakeAgent(logger), logger, false, "MIT-MAGIC-COOKIE-1", "abcd", 0, nil); err != nil {
		t.Fatal(err)
	}

	shutdownWithin(t, forward)
}

func TestOnShutdownTwice(t *testing.T) {
	logger := discardLogger(t)
	forward := agentforward.NewAgentForward(logger)
	if err := forward.NewTCPReverseForwarding(fakeAgent(logger), logger, "127.0.0.1", 8080, nil); err != nil {
		t.Fatal(err)
	}
	if err := forward.NewX11Forwarding(fakeAgent(logger), logger, false, "MIT-MAGIC-COOKIE-1", "abcd", 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := forward.NewForwardTCP(fakeAgent(logger), logger, "127.0.0.1", 22, "127.0.0.1", 1234); err != nil {
		t.Fatal(err)
	}

	shutdownWithin(t, forward)
	shutdownWithin(t, forward)
}

func shutdownWithin(t *testing.T, forward agentforward.AgentForward) {
	t.Helper()
	if !agentforward.ShutdownWithTimeout(forward, 10*time.Second) {
		t.Fatal("OnShutdown did not return in time")
	}
}

// fakeAgent returns a setup callback that starts an in-process agent speaking the agent protocol instead of a
// containerssh-agent process in a container.
func fakeAgent(logger log.Logger) func() (io.Reader, io.Writer, error) {
	return func() (io.Reader, io.Writer, error) {
		fromAgentReader, fromAgentWriter := io.Pipe()
		toAgentReader, toAgentWriter := io.Pipe()
		agent := protocol.NewForwardCtx(toAgentReader, fromAgentWriter, logger)
		go func() {
			connectionType, _, connChan, err := agent.StartClient()
			if err != nil {
				return
			}
			// For direct forwarding the server opens connections; accept them and leave them open so shutdown has to
			// close them. For reverse and X11 forwarding the agent listens until it is told to stop.
			for conn := range connChan {
				if connectionType == protocol.CONNECTION_TYPE_PORT_DIAL {
					_ = conn.Accept()
				}
			}
		}()
		return fromAgentReader, toAgentWriter, nil
	}
}

// discardLogger is used because the forwarding goroutines may log after the test has finished, which a test logger
// does not allow.
func discardLogger(t *testing.T) log.Logger {
	logger, err := log.NewLogger(
		config.LogConfig{
			Level:       config.LogLevelDebug,
			Format:      config.LogFormatText,
			Destination: config.LogDestinationStdout,
			Stdout:      io.Discard,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return logger
}
