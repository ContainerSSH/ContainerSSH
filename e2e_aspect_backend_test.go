package containerssh_test

import (
	"fmt"

	configuration "github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/internal/structutils"
	"github.com/containerssh/libcontainerssh/log"
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
	return fmt.Errorf("Fix this test please")
	/*config.Backend = "kubernetes"
	return nil*/
}

func (k *kubernetesTestingFactor) StartBackingServices(_ configuration.AppConfig, _ log.Logger) error {
	// Assume Kubernetes is already running
	return nil
}

func (k *kubernetesTestingFactor) GetSteps(
	_ configuration.AppConfig,
	_ log.Logger,
) []Step {
	return []Step{}
}

func (k *kubernetesTestingFactor) StopBackingServices(_ configuration.AppConfig, _ log.Logger) error {
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

func (d *dockerTestingFactor) StartBackingServices(_ configuration.AppConfig, _ log.Logger) error {
	// Assume Docker is already running
	return nil
}

func (d *dockerTestingFactor) GetSteps(
	_ configuration.AppConfig,
	_ log.Logger,
) []Step {
	return []Step{}
}

func (d *dockerTestingFactor) StopBackingServices(_ configuration.AppConfig, _ log.Logger) error {
	return nil
}
