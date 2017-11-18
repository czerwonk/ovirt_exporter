package statistic

import (
	"strings"

	"fmt"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/prometheus/client_golang/prometheus"
)

func CollectMetrics(path, prefix string, labelNames, labelValues []string, client *api.ApiClient, ch chan<- prometheus.Metric) error {
	stats := Statistics{}
	err := client.GetAndParse(path, &stats)
	if err != nil {
		return err
	}

	for _, s := range stats.Statistic {
		if s.Kind == "gauge" {
			ch <- convertToMetric(s, prefix, labelNames, labelValues)
		}
	}
	return nil
}

func convertToMetric(s Statistic, prefix string, labelNames, labelValues []string) prometheus.Metric {
	metricName := strings.Replace(s.Name, ".", "_", -1)

	if s.Unit != "none" {
		metricName += "_" + s.Unit
	}

	d := prometheus.NewDesc(fmt.Sprint(prefix, metricName), s.Description, labelNames, nil)

	return metric.MustCreate(d, s.Values.Value.Datum, labelValues)
}
