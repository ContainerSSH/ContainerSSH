package test_test

import (
	"testing"

    "go.containerssh.io/libcontainerssh/internal/test"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestKubernetes(t *testing.T) {
	// Force the test to create a Kubernetes cluster, even if a local config exists.
	t.Setenv("HOME", "/nonexistent")
	t.Setenv("home", "/nonexistent")
	t.Setenv("USERPROFILE", "C:\\nonexistent")
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

// TestKubernetesModeSwitch tests Kubernetes in the real operational mode, where an existing Kubernetes cluster may be
// used.
func TestKubernetesModeSwitch(t *testing.T) {
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
