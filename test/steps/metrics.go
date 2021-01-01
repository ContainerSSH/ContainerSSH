package steps

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func (scenario *Scenario) MetricIsVisible(metric string) error {
	metrics, err := http.Get("http://localhost:9100/metrics")
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(metrics.Body)
	if err != nil {
		return err
	}
	if err := metrics.Body.Close(); err != nil {
		return err
	}
	if !strings.Contains(string(data), metric) {
		return fmt.Errorf("failed to find metric \"%s\" in metrics output:\n%s", metric, data)
	}
	return nil
}
