package test_test

import (
	"os"
	"testing"

	"github.com/containerssh/libcontainerssh/internal/test"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestKubernetes(t *testing.T) {
	// Force the test to create a Kubernetes cluster, even if a local config exists.
	originalHomeLinux := os.Getenv("HOME")
	_ = os.Setenv("HOME", "/nonexistent")
	originalHomePlan9 := os.Getenv("home")
	_ = os.Setenv("home", "/nonexistent")
	originalHomeWindows := os.Getenv("USERPROFILE")
	_ = os.Setenv("USERPROFILE", "C:\\nonexistent")
	t.Cleanup(func() {
		_ = os.Setenv("HOME", originalHomeLinux)
		_ = os.Setenv("home", originalHomePlan9)
		_ = os.Setenv("USERPROFILE", originalHomeWindows)
	})
	kube := test.Kubernetes(t)

	cli, err := kubernetes.NewForConfig(
		&rest.Config{
			Host: kube.Host,
			TLSClientConfig: rest.TLSClientConfig{
				ServerName: kube.ServerName,
				CertData:   []byte(kube.UserCert),
				KeyData:    []byte(kube.UserKey),
				CAData:     []byte(kube.CACert),
			},
		},
	)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client (%v)", err)
	}
	serverVersion, err := cli.ServerVersion()
	if err != nil {
		t.Fatalf("Failed to get server version (%v)", err)
	}

	t.Logf("Server version is %s", serverVersion.String())
}
