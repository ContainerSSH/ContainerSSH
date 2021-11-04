package test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/log"
)

// kubernetesClusterLock allows creating only a single Kubernetes cluster at a time to conserve resources.
var kubernetesClusterLock = &sync.Mutex{}

// Kubernetes launches a Kubernetes-in-Docker cluster for testing purposes. The returned string is the path
// to the kubeconfig file. This function MAY block until enough resources become available to run the Kubernetes
// cluster.
func Kubernetes(t *testing.T) string {
	clusterName := strings.Replace(strings.ToLower(t.Name()), "/", ".", -1)

	tmpHomePath, err := os.MkdirTemp("", "kubeconfig-")
	if err != nil {
		t.Fatalf("failed to create temporary directory for kubeconfig (%v)", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpHomePath)
	})

	kubeConfigPath := filepath.Join(tmpHomePath, ".kube")
	if err := os.MkdirAll(kubeConfigPath, 0700); err != nil {
		t.Fatalf("failed to create kubeconfig path %s (%v)", kubeConfigPath, err)
	}
	kubeConfigFile := filepath.Join(kubeConfigPath, "config")

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(&kubeLogAdapter{t: t}),
	)

	kubernetesClusterLock.Lock()
	t.Cleanup(func() {
		kubernetesClusterLock.Unlock()
	})

	if err := provider.Create(
		clusterName,
		cluster.CreateWithKubeconfigPath(kubeConfigFile),
	); err != nil {
		t.Fatalf("failed to create Kubernetes cluster (%v)", err)
	}
	t.Cleanup(func() {
		if err := provider.Delete(clusterName, kubeConfigFile); err != nil {
			t.Fatalf("failed to remove temporary Kubernetes cluster %s (%v)", clusterName, err)
		}
	})
	return kubeConfigFile
}

type kubeLogAdapter struct {
	t *testing.T
}

func (k kubeLogAdapter) Info(message string) {
	k.t.Log(message)
}

func (k kubeLogAdapter) Infof(format string, args ...interface{}) {
	k.t.Logf(format, args...)
}

func (k kubeLogAdapter) Enabled() bool {
	return true
}

func (k kubeLogAdapter) Warn(message string) {
	k.t.Log(message)
}

func (k kubeLogAdapter) Warnf(format string, args ...interface{}) {
	k.t.Logf(format, args...)
}

func (k kubeLogAdapter) Error(message string) {
	k.t.Log(message)
}

func (k kubeLogAdapter) Errorf(format string, args ...interface{}) {
	k.t.Logf(format, args...)
}

func (k *kubeLogAdapter) V(_ log.Level) log.InfoLogger {
	return k
}
