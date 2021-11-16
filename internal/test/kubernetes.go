package test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/log"
	"sigs.k8s.io/yaml"
)

// kubernetesClusterLock allows creating only a single Kubernetes cluster at a time to conserve resources.
var kubernetesClusterLock = &sync.Mutex{}

// Kubernetes launches a Kubernetes-in-Docker cluster for testing purposes. The returned string is the path
// to the kubeconfig file. This function MAY block until enough resources become available to run the Kubernetes
// cluster. This function MAY return a shared Kubernetes cluster.
func Kubernetes(t *testing.T) KubernetesTestConfiguration {
	t.Helper()
	kubeConfigFile, err := getKubeConfigFromUserHome(t)
	if err != nil {
		t.Logf(
			"Loading kubeconfig from user home directory failed, falling back to launching a KinD cluster (%v)",
			err,
		)
		kubeConfigFile = launchKubernetesCluster(t)
	}

	cfg, err := readKubeConfig(kubeConfigFile)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig from %s (%v)", kubeConfigFile, err)
	}

	userKey, err := base64.StdEncoding.DecodeString(cfg.getUser(t).User.ClientKeyData)
	if err != nil {
		t.Fatalf("Failed to decode client key from kubeconfig %s (%v)", kubeConfigFile, err)
	}
	userCert, err := base64.StdEncoding.DecodeString(cfg.getUser(t).User.ClientCertificateData)
	if err != nil {
		t.Fatalf("Failed to decode client cert from kubeconfig %s (%v)", kubeConfigFile, err)
	}
	caCert, err := base64.StdEncoding.DecodeString(cfg.getCluster(t).Cluster.CertificateAuthorityData)
	if err != nil {
		t.Fatalf("Failed to decode CA cert from kubeconfig %s (%v)", kubeConfigFile, err)
	}

	t.Logf("Kubernetes cluster is now ready for use at %s", cfg.getCluster(t).Cluster.Server)

	hostname := regexp.MustCompile(":.*").ReplaceAllString(
		strings.Replace(cfg.getCluster(t).Cluster.Server, "https://", "", -1),
		"",
	)
	if hostname == "127.0.0.1" {
		hostname = "localhost"
	}
	return KubernetesTestConfiguration{
		Host:       cfg.getCluster(t).Cluster.Server,
		ServerName: hostname,
		CACert:     string(caCert),
		UserCert:   string(userCert),
		UserKey:    string(userKey),
	}
}

func launchKubernetesCluster(t *testing.T) string {
	t.Helper()

	clusterName := strings.Replace(
		strings.Replace(
			strings.ToLower(t.Name()),
			"/",
			".",
			-1,
		),
		"=",
		"-",
		-1,
	)
	if len(clusterName) > 42 {
		clusterName = clusterName[:42]
	}

	tmpHomePath, err := os.MkdirTemp("", "kubeconfig-")
	if err != nil {
		t.Fatalf("failed to create temporary directory for kubeconfig (%v)", err)
	}
	t.Cleanup(
		func() {
			_ = os.RemoveAll(tmpHomePath)
		},
	)

	kubeConfigPath := filepath.Join(tmpHomePath, ".kube")
	if err := os.MkdirAll(kubeConfigPath, 0700); err != nil {
		t.Fatalf("failed to create kubeconfig path %s (%v)", kubeConfigPath, err)
	}
	kubeConfigFile := filepath.Join(kubeConfigPath, "config")

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(&kubeLogAdapter{t: t}),
	)

	kubernetesClusterLock.Lock()
	t.Cleanup(
		func() {
			kubernetesClusterLock.Unlock()
		},
	)

	if err := provider.Create(
		clusterName,
		cluster.CreateWithKubeconfigPath(kubeConfigFile),
	); err != nil {
		t.Fatalf("failed to create Kubernetes cluster (%v)", err)
	}
	t.Cleanup(
		func() {
			t.Logf("Test finished, removing Kubernetes cluster...")
			if err := provider.Delete(clusterName, kubeConfigFile); err != nil {
				t.Fatalf("failed to remove temporary Kubernetes cluster %s (%v)", clusterName, err)
			}
			t.Logf("Kubernetes cluster removed.")
		},
	)
	return kubeConfigFile
}

func getKubeConfigFromUserHome(t *testing.T) (string, error) {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to fetch user home directory (%w)", err)
	}
	kubectlConfig := filepath.Join(home, ".kube", "config")
	if _, err := os.Stat(kubectlConfig); err != nil {
		return "", fmt.Errorf("kubectl configuration does not exist in %s (%w)", kubectlConfig, err)
	}
	return kubectlConfig, nil
}

// KubernetesTestConfiguration holds the credentials to the Kubernetes cluster that can be used for testing.
type KubernetesTestConfiguration struct {
	// Host contains the IP, IP and port, or URL to connect to.
	Host string
	// ServerName contains the host name against which the servers certificate should be validated.
	ServerName string
	// CACert contains the clusters CA certificate in PEM format.
	CACert string
	// UserCert contains the users certificate in PEM format.
	UserCert string
	// UserKey contains the users private key in PEM format.
	UserKey string
}

type kubeLogAdapter struct {
	t *testing.T
}

func (k kubeLogAdapter) Info(message string) {
	k.t.Helper()
	k.t.Log(message)
}

func (k kubeLogAdapter) Infof(format string, args ...interface{}) {
	k.t.Helper()
	k.t.Logf(format, args...)
}

func (k kubeLogAdapter) Enabled() bool {
	k.t.Helper()
	return true
}

func (k kubeLogAdapter) Warn(message string) {
	k.t.Helper()
	k.t.Log(message)
}

func (k kubeLogAdapter) Warnf(format string, args ...interface{}) {
	k.t.Helper()
	k.t.Logf(format, args...)
}

func (k kubeLogAdapter) Error(message string) {
	k.t.Helper()
	k.t.Log(message)
}

func (k kubeLogAdapter) Errorf(format string, args ...interface{}) {
	k.t.Helper()
	k.t.Logf(format, args...)
}

func (k *kubeLogAdapter) V(_ log.Level) log.InfoLogger {
	return k
}

type kubeConfig struct {
	ApiVersion     string              `json:"apiVersion" default:"v1"`
	Clusters       []kubeConfigCluster `json:"clusters"`
	Contexts       []kubeConfigContext `json:"contexts"`
	CurrentContext string              `json:"current-context"`
	Kind           string              `json:"kind" default:"KubernetesConfig"`
	Preferences    map[string]string   `json:"preferences"`
	Users          []kubeConfigUser    `json:"users"`
}

func (k kubeConfig) getContext(t *testing.T) kubeConfigContext {
	t.Helper()
	if k.CurrentContext == "" {
		return k.Contexts[0]
	}
	for _, context := range k.Contexts {
		if context.Name == k.CurrentContext {
			return context
		}
	}
	t.Fatalf("Context %s not found", k.CurrentContext)
	return kubeConfigContext{}
}

func (k kubeConfig) getCluster(t *testing.T) kubeConfigCluster {
	t.Helper()
	context := k.getContext(t)

	for _, cluster := range k.Clusters {
		if cluster.Name == context.Context.Cluster {
			return cluster
		}
	}
	t.Fatalf("Cluster %s not found in kubeconfig", context.Context.Cluster)
	return kubeConfigCluster{}
}

func (k kubeConfig) getUser(t *testing.T) kubeConfigUser {
	t.Helper()
	context := k.getContext(t)

	for _, user := range k.Users {
		if user.Name == context.Context.User {
			return user
		}
	}
	t.Fatalf("User %s not found in kubeconfig", context.Context.Cluster)
	return kubeConfigUser{}
}

type kubeConfigCluster struct {
	Name    string `yaml:"name"`
	Cluster struct {
		CertificateAuthorityData string `json:"certificate-authority-data"`
		Server                   string `json:"server"`
	} `yaml:"cluster"`
}

type kubeConfigContext struct {
	Name    string `json:"name"`
	Context struct {
		Cluster string `json:"cluster"`
		User    string `json:"user"`
	} `yaml:"context"`
}

type kubeConfigUser struct {
	Name string `yaml:"name"`
	User struct {
		ClientCertificateData string `json:"client-certificate-data"`
		ClientKeyData         string `json:"client-key-data"`
		Token                 string `json:"token"`
	} `yaml:"user"`
}

func readKubeConfig(file string) (config kubeConfig, err error) {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
