package containerssh_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/containerssh/libcontainerssh/config"
	"github.com/containerssh/libcontainerssh/log"
	"github.com/containerssh/libcontainerssh/message"
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
	config  config.AppConfig
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

func (m *metricsFactor) ModifyConfiguration(cfg *config.AppConfig) error {
	cfg.Metrics.Enable = m.enabled
	// Change the metrics port because 9100 is often the printer port on desktop OS, which can lead to weird conflicts.
	cfg.Metrics.Listen = "0.0.0.0:9101"
	return nil
}

func (m *metricsFactor) StartBackingServices(
	cfg config.AppConfig,
	_ log.Logger,
) error {
	m.config = cfg
	return nil
}

func (m *metricsFactor) GetSteps(
	cfg config.AppConfig,
	logger log.Logger,
) []Step {
	step := &metricsStep{
		config: cfg,
		logger: logger,
	}
	return []Step{
		{
			`^the "([^"]*)" metric should be visible$`,
			step.TheMetricShouldBeVisible,
		},
	}
}

func (m *metricsFactor) StopBackingServices(_ config.AppConfig, _ log.Logger) error {
	return nil
}

type metricsStep struct {
	config config.AppConfig
	logger log.Logger
}

func (m *metricsStep) TheMetricShouldBeVisible(metricName string) error {
	if !m.config.Metrics.Enable {
		m.logger.Notice(message.NewMessage(message.MTest, "test skipped, metrics not enabled"))
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
