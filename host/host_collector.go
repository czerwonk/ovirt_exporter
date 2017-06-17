package host

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/statistic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// HostCollector collects host statistics from oVirt
type HostCollector struct {
	api       *api.ApiClient
	metrics   []prometheus.Metric
	mutex     sync.Mutex
	retriever *statistic.StatisticMetricRetriever
}

// NewCollector creates a new collector
func NewCollector(c *api.ApiClient) prometheus.Collector {
	r := statistic.NewStatisticMetricRetriever("host", c)
	return &HostCollector{api: c, retriever: r}
}

// Collect implements Prometheus Collector interface
func (c *HostCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.getMetrics() {
		ch <- m
	}
}

// Describe implements Prometheus Collector interface
func (c *HostCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.getMetrics() {
		ch <- m.Desc()
	}
}

func (c *HostCollector) getMetrics() []prometheus.Metric {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.metrics != nil {
		return c.metrics
	}

	c.retrieveMetrics()
	return c.metrics
}

func (c *HostCollector) retrieveMetrics() {
	ressources := make(map[string]string)
	for _, h := range c.getHosts() {
		ressources[h.Id] = h.Name
	}

	c.metrics = c.retriever.RetrieveMetrics(ressources)
}

func (c *HostCollector) getHosts() []Host {
	var hosts Hosts
	err := c.api.GetAndParse("hosts", &hosts)

	if err != nil {
		log.Error(err)
	}

	return hosts.Host
}
