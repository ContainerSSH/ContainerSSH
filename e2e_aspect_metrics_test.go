package containerssh_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/containerssh/configuration"
	"github.com/containerssh/log"
)

func NewMetricsTestingAspect() TestingAspect {
	return &metricsTestingAspect{}
}

type metricsTestingAspect struct {
}

func (m *metricsTestingAspect) String() string {
	return "Metrics"
}

func (m *metricsTestingAspect) Factors() []TestingFactor {
	return []TestingFactor{
		&metricsFactor{
			enabled: true,
			aspect:  m,
		},
		&metricsFactor{
			enabled: false,
			aspect:  m,
		},
	}
}

type metricsFactor struct {
	enabled bool
	config  configuration.AppConfig
	aspect  *metricsTestingAspect
}

func (m *metricsFactor) Aspect() TestingAspect {
	return m.aspect
}

func (m *metricsFactor) String() string {
	if m.enabled {
		return "enabled"
	} else {
		return "disabled"
	}
}

func (m *metricsFactor) ModifyConfiguration(config *configuration.AppConfig) error {
	config.Metrics.Enable = m.enabled
	// Change the metrics port because 9100 is often the printer port on desktop OS, which can lead to weird conflicts.
	config.Metrics.Listen = "0.0.0.0:9101"
	return nil
}

func (m *metricsFactor) StartBackingServices(
	config configuration.AppConfig,
	_ log.Logger,
) error {
	m.config = config
	return nil
}

func (m *metricsFactor) GetSteps(
	config configuration.AppConfig,
	logger log.Logger,
) []Step {
	step := &metricsStep{
		config: config,
		logger: logger,
	}
	return []Step{
		{
			`^the "([^"]*)" metric should be visible$`,
			step.TheMetricShouldBeVisible,
		},
	}
}

func (m *metricsFactor) StopBackingServices(_ configuration.AppConfig, _ log.Logger) error {
	return nil
}

type metricsStep struct {
	config configuration.AppConfig
	logger log.Logger
}

func (m *metricsStep) TheMetricShouldBeVisible(metricName string) error {
	if !m.config.Metrics.Enable {
		m.logger.Notice(log.NewMessage(log.MTest, "test skipped, metrics not enabled"))
		return nil
	}
	metricsResponse, err := http.Get(
		"http://" + strings.Replace(m.config.Metrics.Listen, "0.0.0.0:", "127.0.0.1:", 1) + m.config.Metrics.Path)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(metricsResponse.Body)
	if err != nil {
		return err
	}
	if err := metricsResponse.Body.Close(); err != nil {
		return err
	}
	if !strings.Contains(string(body), "# TYPE "+metricName+" ") {
		return fmt.Errorf("metric %s not found in metrics output:\n%s", metricName, string(body))
	}
	return nil
}
