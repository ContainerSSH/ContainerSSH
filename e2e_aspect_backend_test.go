package containerssh_test

import (
	"github.com/containerssh/configuration"
	"github.com/containerssh/log"
	"github.com/containerssh/structutils"
)

func NewBackendTestingAspect() TestingAspect {
	return &backendTestingAspect{}
}

type backendTestingAspect struct {
}

func (b *backendTestingAspect) String() string {
	return "Backend"
}

func (b *backendTestingAspect) Factors() []TestingFactor {
	return []TestingFactor{
		&kubernetesTestingFactor{
			aspect: b,
		},
		&dockerTestingFactor{
			aspect: b,
		},
	}
}

type kubernetesTestingFactor struct {
	aspect *backendTestingAspect
}

func (k *kubernetesTestingFactor) Aspect() TestingAspect {
	return k.aspect
}

func (k *kubernetesTestingFactor) String() string {
	return "Kubernetes"
}

func (k *kubernetesTestingFactor) ModifyConfiguration(config *configuration.AppConfig) error {
	err := config.Kubernetes.SetConfigFromKubeConfig()
	if err != nil {
		return err
	}
	config.Backend = "kubernetes"
	return nil
}

func (k *kubernetesTestingFactor) StartBackingServices(configuration.AppConfig, log.Logger, log.LoggerFactory) error {
	// Assume Kubernetes is already running
	return nil
}

func (k *kubernetesTestingFactor) GetSteps(
	config configuration.AppConfig,
	logger log.Logger,
	loggerFactory log.LoggerFactory,
) []Step {
	return []Step{}
}

func (k *kubernetesTestingFactor) StopBackingServices(configuration.AppConfig, log.Logger, log.LoggerFactory) error {
	return nil
}


type dockerTestingFactor struct {
	aspect *backendTestingAspect
}

func (d *dockerTestingFactor) Aspect() TestingAspect {
	return d.aspect
}

func (d *dockerTestingFactor) String() string {
	return "Docker"
}

func (d *dockerTestingFactor) ModifyConfiguration(config *configuration.AppConfig) error {
	structutils.Defaults(&config.Docker)
	config.Backend = "docker"
	return nil
}

func (d *dockerTestingFactor) StartBackingServices(configuration.AppConfig, log.Logger, log.LoggerFactory) error {
	// Assume Docker is already running
	return nil
}

func (d *dockerTestingFactor) GetSteps(
	config configuration.AppConfig,
	logger log.Logger,
	loggerFactory log.LoggerFactory,
) []Step {
	return []Step{}
}

func (d *dockerTestingFactor) StopBackingServices(configuration.AppConfig, log.Logger, log.LoggerFactory) error {
	return nil
}



