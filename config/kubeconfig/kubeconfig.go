package kubeconfig

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type KubectlConfig struct {
	ApiVersion     string            `yaml:"apiVersion" default:"v1"`
	Clusters       []Cluster         `yaml:"clusters"`
	Contexts       []Context         `yaml:"contexts"`
	CurrentContext string            `yaml:"current-context"`
	Kind           string            `yaml:"kind" default:"Config"`
	Preferences    map[string]string `yaml:"preferences"`
	Users          []User            `yaml:"users"`
}

type Cluster struct {
	Name    string `yaml:"name"`
	Cluster struct {
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
		Server                   string `yaml:"server"`
	} `yaml:"cluster"`
}

type Context struct {
	Name    string `yaml:"name"`
	Context struct {
		Cluster string `yaml:"cluster"`
		User    string `yaml:"user"`
	} `yaml:"context"`
}

type User struct {
	Name string `yaml:"name"`
	User struct {
		ClientCertificateData string `yaml:"client-certificate-data"`
		ClientKeyData         string `yaml:"client-key-data"`
	} `yaml:"user"`
}

func Read(file string) KubectlConfig {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	config := &KubectlConfig{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return *config
}
