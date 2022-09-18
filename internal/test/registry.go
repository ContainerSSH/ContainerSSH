package test

import (
    _ "embed"
    "fmt"
    "net/http"
    "strings"
    "testing"
    "time"

    "golang.org/x/crypto/bcrypt"
)

//go:embed registry/Dockerfile
var registryDockerfile []byte

//go:embed registry/config.yml
var registryConfig []byte

// Registry creates a local Docker registry that can be used for pull tests, optionally with authentication.
func Registry(t *testing.T, auth bool) RegistryHelper {
    files := map[string][]byte{
        "Dockerfile": registryDockerfile,
        "config.yml": registryConfig,
    }
    var usernamePtr *string
    var passwordPtr *string
    if auth {
        username := "test"
        password := generateRandomString(16)
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 0)
        if err != nil {
            t.Fatalf("Failed to generate password hash (%v)", err)
        }
        files["config.yml"] = []byte(strings.TrimSpace(string(registryConfig)) + "\n" + `auth:
  htpasswd:
    realm: Registry Realm
    path: /etc/docker/registry/htpasswd`)
        files["htpasswd"] = []byte(fmt.Sprintf("%s:%s", username, hashedPassword))
        files["Dockerfile"] = append(registryDockerfile, []byte("\nCOPY htpasswd /etc/docker/registry/htpasswd")...)
        usernamePtr = &username
        passwordPtr = &password
    }
    container := containerFromBuild(
        t,
        "test-registry",
        files,
        nil,
        nil,
        map[string]string{
            "5000/tcp": "",
        },
    )
    port := container.port("5000/tcp")

    for i := 0; i < 30; i++ {
        time.Sleep(time.Second)
        t.Logf("Waiting for the registry to come up...")
        req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/v2/", port), nil)
        if err != nil {
            t.Fatal(err)
        }
        if auth {
            req.SetBasicAuth(*usernamePtr, *passwordPtr)
        }

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            t.Logf("Registry is not up yet (%v)", err)
            continue
        }
        _ = resp.Body.Close()
        if resp.StatusCode != 200 {
            t.Logf("Registry returned invalid status code (%d)", resp.StatusCode)
            continue
        }
        return &registryHelper{
            port:     port,
            username: usernamePtr,
            password: passwordPtr,
        }
    }

    t.Fatalf("Registry failed to come up in 30 seconds.")
    return nil
}

type RegistryHelper interface {
    Port() int
    Username() *string
    Password() *string
}

type registryHelper struct {
    port     int
    username *string
    password *string
}

func (r registryHelper) Port() int {
    return r.port
}

func (r registryHelper) Username() *string {
    return r.username
}

func (r registryHelper) Password() *string {
    return r.password
}
