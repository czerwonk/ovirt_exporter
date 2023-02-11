// SPDX-License-Identifier: MIT

package collector

import (
	"github.com/czerwonk/ovirt_api/api"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

func NewContext(tracer trace.Tracer, client *api.Client) *CollectorContext {
	return &CollectorContext{
		tracer: tracer,
		client: &clientTracingAdapter{
			client: client,
			tracer: tracer,
		},
	}
}

type CollectorContext struct {
	tracer trace.Tracer
	client *clientTracingAdapter
	ch     chan<- prometheus.Metric
}

func (c *CollectorContext) Clone() *CollectorContext {
	return &CollectorContext{
		tracer: c.tracer,
		client: c.client,
	}
}

func (c *CollectorContext) Tracer() trace.Tracer {
	return c.tracer
}

func (c *CollectorContext) Client() Client {
	return c.client
}

func (c *CollectorContext) MetricsCh() chan<- prometheus.Metric {
	return c.ch
}

func (c *CollectorContext) SetMetricsCh(ch chan<- prometheus.Metric) {
	c.ch = ch
}
