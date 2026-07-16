package http_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.containerssh.io/containerssh/config"
	containersshhttp "go.containerssh.io/containerssh/http"
	"go.containerssh.io/containerssh/log"
)

func TestClientReusesHTTPConnection(t *testing.T) {
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte("{}"))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	defer server.Close()

	client, err := containersshhttp.NewClient(
		config.HTTPClientConfiguration{
			URL:     server.URL,
			Timeout: time.Second,
		},
		log.NewTestLogger(t),
	)
	if err != nil {
		t.Fatalf("failed to create HTTP client: %v", err)
	}

	for request := 0; request < 2; request++ {
		response := struct{}{}
		if status, err := client.Get("", &response); err != nil {
			t.Fatalf("request %d failed: %v", request+1, err)
		} else if status != http.StatusOK {
			t.Fatalf("request %d returned status %d", request+1, status)
		}
	}

	if connections.Load() != 1 {
		t.Fatalf("expected requests to reuse one connection, got %d connections", connections.Load())
	}
}
