// SPDX-License-Identifier: MIT

package host

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/czerwonk/ovirt_exporter/pkg/cluster"
	"github.com/czerwonk/ovirt_exporter/pkg/collector"
	"github.com/czerwonk/ovirt_exporter/pkg/metric"
	"github.com/czerwonk/ovirt_exporter/pkg/network"
	"github.com/czerwonk/ovirt_exporter/pkg/statistic"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const prefix = "ovirt_host_"

var (
	upDesc               *prometheus.Desc
	cpuCoresDesc         *prometheus.Desc
	cpuSocketsDesc       *prometheus.Desc
	cpuThreadsDesc       *prometheus.Desc
	cpuSpeedDesc         *prometheus.Desc
	memoryDesc           *prometheus.Desc
	labelNames           []string
	hostMaintenanceRegex *regexp.Regexp
)

func init() {
	labelNames = []string{"name", "cluster"}
	upDesc = prometheus.NewDesc(prefix+"up", "Host status is up (1) or not (0) or on maintenance (2)", labelNames, nil)
	cpuCoresDesc = prometheus.NewDesc(prefix+"cpu_cores", "Number of CPU cores assigned", labelNames, nil)
	cpuSocketsDesc = prometheus.NewDesc(prefix+"cpu_sockets", "Number of sockets", labelNames, nil)
	cpuThreadsDesc = prometheus.NewDesc(prefix+"cpu_threads", "Number of threads", labelNames, nil)
	cpuSpeedDesc = prometheus.NewDesc(prefix+"cpu_speed_hertz", "CPU speed in hertz", labelNames, nil)
	memoryDesc = prometheus.NewDesc(prefix+"memory_installed_bytes", "Memory installed in bytes", labelNames, nil)
	hostMaintenanceRegex = regexp.MustCompile(`maintenance|installing`)
}

// HostCollector collects host statistics from oVirt
type HostCollector struct {
	collectDuration prometheus.Observer
	cc              *collector.CollectorContext
	metrics         []prometheus.Metric
	collectNetwork  bool
	mutex           sync.Mutex
	rootCtx         context.Context
}

// NewCollector creates a new collector
func NewCollector(ctx context.Context, cc *collector.CollectorContext, collectNetwork bool, collectDuration prometheus.Observer) prometheus.Collector {
	return &HostCollector{
		rootCtx:         ctx,
		cc:              cc,
		collectNetwork:  collectNetwork,
		collectDuration: collectDuration}
}

// Collect implements Prometheus Collector interface
func (c *HostCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, span := c.cc.Tracer().Start(c.rootCtx, "HostCollector.Collect")
	defer span.End()

	for _, m := range c.getMetrics(ctx, span) {
		ch <- m
	}
}

// Describe implements Prometheus Collector interface
func (c *HostCollector) Describe(ch chan<- *prometheus.Desc) {
	ctx, span := c.cc.Tracer().Start(c.rootCtx, "HostCollector.Describe")
	defer span.End()

	for _, m := range c.getMetrics(ctx, span) {
		ch <- m.Desc()
	}
}

func (c *HostCollector) getMetrics(ctx context.Context, span trace.Span) []prometheus.Metric {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.metrics != nil {
		return c.metrics
	}

	c.retrieveMetrics(ctx, span)
	return c.metrics
}

func (c *HostCollector) retrieveMetrics(ctx context.Context, span trace.Span) {
	timer := prometheus.NewTimer(c.collectDuration)
	defer timer.ObserveDuration()

	h := Hosts{}
	err := c.cc.Client().GetAndParse(ctx, "hosts", &h)
	if err != nil {
		c.cc.HandleError(err, span)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(h.Hosts))

	ch := make(chan prometheus.Metric)
	c.cc.SetMetricsCh(ch)
	for _, h := range h.Hosts {
		go c.collectForHost(ctx, h, wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for m := range ch {
		c.metrics = append(c.metrics, m)
	}
}

func (c *HostCollector) collectForHost(ctx context.Context, host Host, wg *sync.WaitGroup) {
	ctx, span := c.cc.Tracer().Start(ctx, "HostCollector.CollectForHost", trace.WithAttributes(
		attribute.String("host_name", host.Name),
		attribute.String("host_id", host.ID),
	))
	defer span.End()

	defer wg.Done()

	h := &host
	l := []string{h.Name, cluster.Name(ctx, h.Cluster.ID, c.cc.Client())}

	c.cc.RecordMetrics(
		c.upMetric(h, l),
		metric.MustCreate(memoryDesc, float64(host.Memory), l),
	)
	c.collectCPUMetrics(h, l)

	statPath := fmt.Sprintf("hosts/%s/statistics", host.ID)
	statistic.CollectMetrics(ctx, statPath, prefix, labelNames, l, c.cc)

	if c.collectNetwork {
		networkPath := fmt.Sprintf("hosts/%s/nics", host.ID)
		network.CollectMetricsForHost(ctx, networkPath, prefix, labelNames, l, c.cc)
	}
}

func (c *HostCollector) collectCPUMetrics(host *Host, l []string) {
	topo := host.CPU.Topology

	c.cc.RecordMetrics(
		metric.MustCreate(cpuCoresDesc, float64(topo.Cores), l),
		metric.MustCreate(cpuThreadsDesc, float64(topo.Threads), l),
		metric.MustCreate(cpuSocketsDesc, float64(topo.Sockets), l),
		metric.MustCreate(cpuSpeedDesc, float64(host.CPU.Speed*1e6), l),
	)
}

func (c *HostCollector) upMetric(host *Host, labelValues []string) prometheus.Metric {
	var status float64
	host_status := host.Status

	if host_status == "up" {
		status = 1
	} else if hostMaintenanceRegex.MatchString(host_status) {
		status = 2
	}

	return metric.MustCreate(upDesc, status, labelValues)
}
