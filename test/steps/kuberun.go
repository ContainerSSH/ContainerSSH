package steps

import (
	"encoding/base64"
	"fmt"
	"github.com/janoszen/containerssh/config"
	"github.com/janoszen/containerssh/config/kubeconfig"
	"log"
	"os/user"
	"path/filepath"
	"strings"
)

func (scenario *Scenario) ConfigureKubernetes(username string) error {
	if scenario.ConfigServer == nil {
		return fmt.Errorf("config server is not running")
	}

	configuration := &config.AppConfig{}
	configuration.Backend = config.BackendKubeRun

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	kubectlConfig := kubeconfig.Read(filepath.Join(usr.HomeDir, ".kube", "config"))
	currentContext := kubectlConfig.CurrentContext
	var context *kubeconfig.Context
	for _, ctx := range kubectlConfig.Contexts {
		if ctx.Name == currentContext {
			context = &ctx
			break
		}
	}
	if context == nil {
		return fmt.Errorf("failed to find current context in kubeconfig")
	}
	userName := context.Context.User
	clusterName := context.Context.Cluster

	var kubeConfigUser *kubeconfig.User
	for _, u := range kubectlConfig.Users {
		if u.Name == userName {
			kubeConfigUser = &u
			break
		}
	}
	if kubeConfigUser == nil {
		return fmt.Errorf("failed to find user in kubeconfig")
	}

	var kubeConfigCluster *kubeconfig.Cluster
	for _, c := range kubectlConfig.Clusters {
		if c.Name == clusterName {
			kubeConfigCluster = &c
			break
		}
	}
	if kubeConfigCluster == nil {
		return fmt.Errorf("failed to find cluster in kubeconfig")
	}

	configuration.KubeRun.Connection.Host = strings.Replace(
		kubeConfigCluster.Cluster.Server,
		"https://",
		"",
		1,
	)
	decodedCa, err := base64.StdEncoding.DecodeString(
		kubeConfigCluster.Cluster.CertificateAuthorityData,
	)
	if err != nil {
		return err
	}
	configuration.KubeRun.Connection.CAData = string(decodedCa)

	decodedKey, err := base64.StdEncoding.DecodeString(
		kubeConfigUser.User.ClientKeyData,
	)
	if err != nil {
		return err
	}
	configuration.KubeRun.Connection.KeyData = string(decodedKey)

	decodedCert, err := base64.StdEncoding.DecodeString(
		kubeConfigUser.User.ClientCertificateData,
	)
	if err != nil {
		return err
	}
	configuration.KubeRun.Connection.CertData = string(decodedCert)

	scenario.ConfigServer.SetUserConfig(username, configuration)
	return nil
}
