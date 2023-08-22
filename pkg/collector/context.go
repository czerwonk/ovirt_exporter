// SPDX-License-Identifier: MIT

package collector

import (
	"github.com/czerwonk/ovirt_api/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
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

// Tracer returns the configured tracer
func (c *CollectorContext) Tracer() trace.Tracer {
	return c.tracer
}

// Client returns the client to query the API
func (c *CollectorContext) Client() Client {
	return c.client
}

// SetMetricsCh sets the metrics channel observed by the collector
func (c *CollectorContext) SetMetricsCh(ch chan<- prometheus.Metric) {
	c.ch = ch
}

// RecordMetrics returns the collected metrics to the collector
func (c *CollectorContext) RecordMetrics(metrics ...prometheus.Metric) {
	for _, m := range metrics {
		c.ch <- m
	}
}

// HandleError handles an error
func (c *CollectorContext) HandleError(err error, span trace.Span) {
	logrus.Error(err)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
