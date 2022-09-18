package test_test

import (
    "fmt"
    "net/http"
    "testing"

    "go.containerssh.io/libcontainerssh/internal/test"
)

func TestRegistry(t *testing.T) {
    registry := test.Registry(t, false)

    resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v2/_catalog", registry.Port()))
    if err != nil {
        t.Fatal(err)
    }
    _ = resp.Body.Close()
    if resp.StatusCode != 200 {
        t.Fatalf("Invalid status code received from registry: %d", resp.StatusCode)
    }
}

func TestRegistryAuthenticated(t *testing.T) {
    registry := test.Registry(t, true)

    resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v2/_catalog", registry.Port()))
    if err != nil {
        t.Fatal(err)
    }
    _ = resp.Body.Close()
    if resp.StatusCode != 401 {
        t.Fatalf("Invalid status code received from registry without authentication: %d", resp.StatusCode)
    }

    req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/v2/_catalog", registry.Port()), nil)
    if err != nil {
        t.Fatal(err)
    }
    req.SetBasicAuth(*registry.Username(), *registry.Password())

    resp, err = http.DefaultClient.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    _ = resp.Body.Close()
    if resp.StatusCode != 200 {
        t.Fatalf("Invalid status code received from registry with authentication: %d", resp.StatusCode)
    }
}
