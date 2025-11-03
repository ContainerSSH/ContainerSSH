package docker //nolint:testpackage

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
    "go.containerssh.io/containerssh/config"
    "go.containerssh.io/containerssh/internal/geoip/dummy"
    "go.containerssh.io/containerssh/internal/metrics"
    "go.containerssh.io/containerssh/internal/structutils"
    "go.containerssh.io/containerssh/internal/test"
    "go.containerssh.io/containerssh/log"
)

func TestPullImageAuthenticated(t *testing.T) {
    registry := test.Registry(t, true)
    metricsCollector := metrics.New(dummy.New())
    clientFactory := &dockerV20ClientFactory{
        backendFailuresMetric: metricsCollector.MustCreateCounter("backend-failures", "requests", ""),
        backendRequestsMetric: metricsCollector.MustCreateCounter("backend-failures", "requests", ""),
    }
    ctx := context.Background()

    t.Run("unauthenticated", func(t *testing.T) {
        cfg := config.DockerConfig{}
        structutils.Defaults(&cfg)
        cfg.Execution.ContainerConfig.Image = fmt.Sprintf("localhost:%d/containerssh/guest-image", registry.Port())

        logger := log.NewTestLogger(t)
        client, err := clientFactory.get(ctx, cfg, logger)
        if err != nil {
            t.Fatal(err)
        }

        pullCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
        defer cancel()
        if err := client.pullImage(pullCtx); err == nil {
            t.Fatalf("Pulling without credentials didn't fail.")
        }
    })
    t.Run("authenticated", func(t *testing.T) {
        cfg := config.DockerConfig{}
        structutils.Defaults(&cfg)
        cfg.Execution.ContainerConfig.Image = fmt.Sprintf("localhost:%d/containerssh/agent", registry.Port())
        cfg.Execution.Auth = &registry.AuthConfig{
            Username:      *registry.Username(),
            Password:      *registry.Password(),
            Email:         "noreply@containerssh.io",
            ServerAddress: fmt.Sprintf("localhost:%d", registry.Port()),
        }

        logger := log.NewTestLogger(t)
        client, err := clientFactory.get(ctx, cfg, logger)
        if err != nil {
            t.Fatal(err)
        }

        pullCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
        defer cancel()
        if err := client.pullImage(pullCtx); err != nil {
            t.Fatalf("Pulling with credentials failed (%v).", err)
        }
    })
}
