package test_test

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/containerssh/libcontainerssh/internal/test"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func TestKubernetes(t *testing.T) {
	kube := test.Kubernetes(t)

	cfg, err := readKubeConfig(kube)
	if err != nil {
		t.Fatalf("Failed to read kubeconfig from %s (%v)", kube, err)
	}

	userKey, err := base64.StdEncoding.DecodeString(cfg.getUser(t).User.ClientKeyData)
	if err != nil {
		t.Fatalf("Failed to decode client key from kubeconfig %s (%v)", kube, err)
	}
	userCert, err := base64.StdEncoding.DecodeString(cfg.getUser(t).User.ClientCertificateData)
	if err != nil {
		t.Fatalf("Failed to decode client cert from kubeconfig %s (%v)", kube, err)
	}
	caCert, err := base64.StdEncoding.DecodeString(cfg.getCluster(t).Cluster.CertificateAuthorityData)
	if err != nil {
		t.Fatalf("Failed to decode CA cert from kubeconfig %s (%v)", kube, err)
	}

	cli, err := kubernetes.NewForConfig(&rest.Config{
		Host: cfg.getCluster(t).Cluster.Server,
		TLSClientConfig: rest.TLSClientConfig{
			ServerName: "kubernetes.default",
			CertData:   userCert,
			KeyData:    userKey,
			CAData:     caCert,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client (%v)", err)
	}
	serverVersion, err := cli.ServerVersion()
	if err != nil {
		t.Fatalf("Failed to get server version (%v)", err)
	}

	t.Logf("Server version is %s", serverVersion.String())
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
