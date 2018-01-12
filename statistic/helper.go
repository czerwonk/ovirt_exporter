package statistic

import (
	"strings"

	"fmt"

	"github.com/czerwonk/ovirt_api/api"
	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// CollectMetrics collects metrics by statics returned by a given url
func CollectMetrics(path, prefix string, labelNames, labelValues []string, client *api.Client, ch chan<- prometheus.Metric) {
	stats := Statistics{}
	err := client.GetAndParse(path, &stats)
	if err != nil {
		log.Errorln(err)
	}

	for _, s := range stats.Statistic {
		if s.Kind == "gauge" || s.Kind == "counter" {
			ch <- convertToMetric(s, prefix, labelNames, labelValues)
		}
	}
}

func convertToMetric(s Statistic, prefix string, labelNames, labelValues []string) prometheus.Metric {
	metricName := strings.Replace(s.Name, ".", "_", -1)

	if s.Unit != "none" {
		metricName += "_" + s.Unit
	}

	d := prometheus.NewDesc(fmt.Sprint(prefix, metricName), s.Description, labelNames, nil)

	return metric.MustCreate(d, s.Values.Value.Datum, labelValues)
}
