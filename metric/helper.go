package metric

import "github.com/prometheus/client_golang/prometheus"

func MustCreate(desc *prometheus.Desc, v float64, labelValues []string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(v), labelValues...)
}
