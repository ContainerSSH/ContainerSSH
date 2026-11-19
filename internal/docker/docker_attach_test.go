package docker //nolint:testpackage

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/geoip/dummy"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/internal/structutils"
	"go.containerssh.io/containerssh/log"
)

// TestSessionModeLogDrivers checks that a session mode container can be attached to, started, and its output read
// for every log driver, including "none", which cannot be read back with docker logs.
func TestSessionModeLogDrivers(t *testing.T) {
	t.Parallel()

	metricsCollector := metrics.New(dummy.New())
	clientFactory := &dockerV20ClientFactory{
		backendFailuresMetric: metricsCollector.MustCreateCounter("backend-failures", "requests", ""),
		backendRequestsMetric: metricsCollector.MustCreateCounter("backend-requests", "requests", ""),
	}

	for _, logDriver := range []string{"", "json-file", "local", "none"} {
		logDriver := logDriver
		name := logDriver
		if name == "" {
			name = "daemon-default"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := config.DockerConfig{}
			structutils.Defaults(&cfg)
			cfg.Execution.Mode = config.DockerExecutionModeSession
			cfg.Execution.HostConfig = &container.HostConfig{}
			if logDriver != "" {
				cfg.Execution.HostConfig.LogConfig = container.LogConfig{Type: logDriver}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			dc, err := clientFactory.get(ctx, cfg, log.NewTestLogger(t))
			if err != nil {
				t.Fatal(err)
			}
			hasImage, err := dc.hasImage(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if !hasImage {
				if err := dc.pullImage(ctx); err != nil {
					t.Fatal(err)
				}
			}

			tty := false
			cnt, err := dc.createContainer(ctx, map[string]string{}, map[string]string{}, &tty, []string{"echo", "hello"})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				removeCtx, removeCancel := context.WithTimeout(context.Background(), time.Minute)
				defer removeCancel()
				_ = cnt.remove(removeCtx)
			}()

			// A failed attach is retried every 10 seconds until the context ends, so give it a short deadline.
			attachCtx, attachCancel := context.WithTimeout(ctx, 5*time.Second)
			defer attachCancel()
			exec, err := cnt.attach(attachCtx)
			if err != nil {
				t.Fatalf("failed to attach to the container with log driver %q: %v", logDriver, err)
			}
			if err := cnt.start(ctx); err != nil {
				t.Fatal(err)
			}

			stdout := &syncBuffer{}
			stderr := &syncBuffer{}
			exitCode := make(chan int, 1)
			exec.run(
				strings.NewReader(""),
				stdout,
				stderr,
				func() error { return nil },
				func(exitStatus int) { exitCode <- exitStatus },
			)

			select {
			case code := <-exitCode:
				if code != 0 {
					t.Fatalf("unexpected exit code %d, stderr: %s", code, stderr.String())
				}
			case <-ctx.Done():
				t.Fatal("timeout waiting for the container to exit")
			}
			if got := stdout.String(); got != "hello\n" {
				t.Fatalf("unexpected container output %q", got)
			}
		})
	}
}

// syncBuffer is a bytes.Buffer that is safe to write from the exec goroutines while the test reads it.
type syncBuffer struct {
	lock sync.Mutex
	buf  bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.buf.String()
}
