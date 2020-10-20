package metrics

import (
	"sort"
	"strings"
)

type Metric struct {
	Name   string
	Labels map[string]string
}

func (metric *Metric) ToString() string {
	var labelList []string

	keys := make([]string, 0, len(metric.Labels))
	for k := range metric.Labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		labelList = append(labelList, k+"=\""+metric.Labels[k]+"\"")
	}

	var labels string
	if len(labelList) > 0 {
		labels = "{" + strings.Join(labelList, ",") + "}"
	} else {
		labels = ""
	}

	return metric.Name + labels
}
