package statistic

import (
	"fmt"
	"strings"

	"github.com/czerwonk/ovirt_api/api"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// CollectMetrics collects metrics by statics returned by a given url
func CollectMetrics(path, prefix string, labelNames, labelValues []string, client *api.Client, ch chan<- prometheus.Metric) {
	stats := Statistics{}
	err := client.GetAndParse(path, &stats)
	if err != nil {
		log.Errorln(err)
	}

	for _, s := range stats.Statistic {
		if s.Type != "decimal" && s.Type != "integer" {
			continue
		}
		switch s.Kind {
		case "gauge":
			ch <- convertToMetric(s, prefix, labelNames, labelValues, prometheus.GaugeValue)
		case "counter":
			ch <- convertToMetric(s, prefix, labelNames, labelValues, prometheus.CounterValue)
		}
	}
}

func convertToMetric(s Statistic, prefix string, labelNames, labelValues []string, valueType prometheus.ValueType) prometheus.Metric {
	metricName := strings.Replace(s.Name, ".", "_", -1)

	if s.Unit != "none" {
		metricName += "_" + s.Unit
	}

	if valueType == prometheus.CounterValue {
		// Suffix counter metrics with '_total' to follow Prometheus best practices.
		metricName = strings.ReplaceAll(metricName, "_total", "")
		metricName = metricName + "_total"
	}
	d := prometheus.NewDesc(fmt.Sprint(prefix, metricName), s.Description, labelNames, nil)

	return prometheus.MustNewConstMetric(d, valueType, float64(s.Values.Value.Datum), labelValues...)
}
