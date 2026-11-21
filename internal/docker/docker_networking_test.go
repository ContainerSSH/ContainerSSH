package docker //nolint:testpackage

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/internal/geoip/dummy"
	"go.containerssh.io/containerssh/internal/metrics"
	"go.containerssh.io/containerssh/log"
	"gopkg.in/yaml.v3"
)

// fakeDockerDaemon is a minimal fake of the Docker Engine API that records the bodies of
// container create requests.
type fakeDockerDaemon struct {
	server *httptest.Server
	mutex  sync.Mutex
	bodies [][]byte
}

func newFakeDockerDaemon(t *testing.T) *fakeDockerDaemon {
	daemon := &fakeDockerDaemon{}
	daemon.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/_ping":
			w.Header().Set("Api-Version", "1.47")
			w.Header().Set("Docker-Experimental", "false")
			w.Header().Set("Ostype", "linux")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/containers/create"):
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			daemon.mutex.Lock()
			daemon.bodies = append(daemon.bodies, body)
			daemon.mutex.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"Id":"0123456789abcdef0123456789abcdef","Warnings":[]}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	t.Cleanup(daemon.server.Close)
	return daemon
}

// lastCreateBody returns the decoded body of the last container create request.
func (d *fakeDockerDaemon) lastCreateBody(t *testing.T) map[string]interface{} {
	t.Helper()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if len(d.bodies) == 0 {
		t.Fatal("no container create request was received by the fake docker daemon")
	}
	var body map[string]interface{}
	if err := json.Unmarshal(d.bodies[len(d.bodies)-1], &body); err != nil {
		t.Fatalf("failed to decode container create body %q: %v", d.bodies[len(d.bodies)-1], err)
	}
	return body
}

// networkingEndpointNetworkID extracts NetworkingConfig.EndpointsConfig[name].NetworkID from a decoded
// container create request body.
func networkingEndpointNetworkID(t *testing.T, body map[string]interface{}, name string) (string, bool) {
	t.Helper()
	networkingConfig, ok := body["NetworkingConfig"].(map[string]interface{})
	if !ok {
		return "", false
	}
	endpoints, ok := networkingConfig["EndpointsConfig"].(map[string]interface{})
	if !ok {
		return "", false
	}
	endpoint, ok := endpoints[name].(map[string]interface{})
	if !ok {
		return "", false
	}
	networkID, _ := endpoint["NetworkID"].(string)
	return networkID, true
}

// createContainerOnFakeDaemon builds a dockerV20Client talking to the fake daemon and creates a container.
func createContainerOnFakeDaemon(t *testing.T, cfg config.DockerConfig, daemon *fakeDockerDaemon) {
	t.Helper()
	cli, err := client.NewClientWithOpts(
		client.WithHost(daemon.server.URL),
		client.WithHTTPClient(daemon.server.Client()),
		client.WithVersion("1.47"),
	)
	if err != nil {
		t.Fatal(err)
	}
	metricsCollector := metrics.New(dummy.New())
	dockerClient := &dockerV20Client{
		config:                cfg,
		dockerClient:          cli,
		logger:                log.NewTestLogger(t),
		backendFailuresMetric: metricsCollector.MustCreateCounter("backend-failures", "requests", ""),
		backendRequestsMetric: metricsCollector.MustCreateCounter("backend-requests", "requests", ""),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := dockerClient.createContainer(ctx, map[string]string{}, map[string]string{}, nil, nil); err != nil {
		t.Fatalf("failed to create container: %v", err)
	}
}

// TestCreateContainerSendsConfiguredNetworkingConfig is a regression test for issue #639: when the
// configuration places networkingconfig.endpointsconfig.<name>.networkid inside the "container" section,
// the container create request sent to the Docker daemon must contain that NetworkingConfig, otherwise the
// container is created on the default bridge network.
func TestCreateContainerSendsConfiguredNetworkingConfig(t *testing.T) {
	// The docker section of the config.yaml from issue #639.
	issueConfig := `
connection:
  host: "unix:///var/run/docker.sock"
execution:
  container:
    image: custom-agent/containerssh-agent:v0
    networkingconfig:
      endpointsconfig:
        administrator_default:
          networkid: "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1"
  network:
    name: "administrator_default"
  mode: session
`
	cfg := config.DockerConfig{}
	if err := yaml.Unmarshal([]byte(issueConfig), &cfg); err != nil {
		t.Fatal(err)
	}

	daemon := newFakeDockerDaemon(t)
	createContainerOnFakeDaemon(t, cfg, daemon)

	body := daemon.lastCreateBody(t)
	networkID, ok := networkingEndpointNetworkID(t, body, "administrator_default")
	if !ok {
		t.Fatalf(
			"container create request body does not contain NetworkingConfig.EndpointsConfig.administrator_default: %s",
			mustMarshal(t, body),
		)
	}
	if networkID != "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1" {
		t.Errorf("unexpected NetworkID in container create request: %q", networkID)
	}
}

// TestCreateContainerSendsNetworkKeyNetworkingConfig ensures the documented "network" key keeps being sent.
func TestCreateContainerSendsNetworkKeyNetworkingConfig(t *testing.T) {
	issueConfig := `
connection:
  host: "unix:///var/run/docker.sock"
execution:
  container:
    image: custom-agent/containerssh-agent:v0
  network:
    endpointsconfig:
      administrator_default:
        networkid: "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1"
  mode: session
`
	cfg := config.DockerConfig{}
	if err := yaml.Unmarshal([]byte(issueConfig), &cfg); err != nil {
		t.Fatal(err)
	}

	daemon := newFakeDockerDaemon(t)
	createContainerOnFakeDaemon(t, cfg, daemon)

	body := daemon.lastCreateBody(t)
	networkID, ok := networkingEndpointNetworkID(t, body, "administrator_default")
	if !ok {
		t.Fatalf(
			"container create request body does not contain NetworkingConfig.EndpointsConfig.administrator_default: %s",
			mustMarshal(t, body),
		)
	}
	if networkID != "9e8ecb7bab9a9c0e6cf957b42827c2df759246f89a33e5ef8815442880eabef1" {
		t.Errorf("unexpected NetworkID in container create request: %q", networkID)
	}
}

// TestCreateContainerDefaultNetworkWhenUnconfigured ensures the default behavior is preserved: when no
// networking config is set, no NetworkingConfig endpoints are sent and the daemon applies its defaults.
func TestCreateContainerDefaultNetworkWhenUnconfigured(t *testing.T) {
	issueConfig := `
connection:
  host: "unix:///var/run/docker.sock"
execution:
  container:
    image: custom-agent/containerssh-agent:v0
  mode: session
`
	cfg := config.DockerConfig{}
	if err := yaml.Unmarshal([]byte(issueConfig), &cfg); err != nil {
		t.Fatal(err)
	}

	daemon := newFakeDockerDaemon(t)
	createContainerOnFakeDaemon(t, cfg, daemon)

	body := daemon.lastCreateBody(t)
	if _, ok := networkingEndpointNetworkID(t, body, "administrator_default"); ok {
		t.Errorf(
			"container create request unexpectedly contains NetworkingConfig.EndpointsConfig.administrator_default: %s",
			mustMarshal(t, body),
		)
	}
	networkingConfig, ok := body["NetworkingConfig"].(map[string]interface{})
	if ok {
		if endpoints, ok := networkingConfig["EndpointsConfig"].(map[string]interface{}); ok && len(endpoints) > 0 {
			t.Errorf(
				"container create request unexpectedly contains NetworkingConfig.EndpointsConfig endpoints: %s",
				mustMarshal(t, body),
			)
		}
	}
}

func mustMarshal(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
