// SPDX-License-Identifier: MIT

package statistic

import (
	"context"
	"fmt"
	"strings"

	"github.com/czerwonk/ovirt_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CollectMetrics collects metrics by statics returned by a given url
func CollectMetrics(ctx context.Context, path, prefix string, labelNames, labelValues []string, cc *collector.CollectorContext) {
	ctx, span := cc.Tracer().Start(ctx, "Statistic.CollectMetrics", trace.WithAttributes(
		attribute.String("prefix", prefix),
	))
	defer span.End()

	stats := Statistics{}
	err := cc.Client().GetAndParse(ctx, path, &stats)
	if err != nil {
		cc.HandleError(err, span)
	}

	for _, s := range stats.Statistic {
		if s.Type != "decimal" && s.Type != "integer" {
			continue
		}
		switch s.Kind {
		case "gauge":
			cc.RecordMetrics(convertToMetric(s, prefix, labelNames, labelValues, prometheus.GaugeValue))
		case "counter":
			cc.RecordMetrics(convertToMetric(s, prefix, labelNames, labelValues, prometheus.CounterValue))
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
