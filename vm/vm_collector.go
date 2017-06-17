package vm

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/statistic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// VmCollector collects virtual machine statistics from oVirt
type VmCollector struct {
	api       *api.ApiClient
	metrics   []prometheus.Metric
	mutex     sync.Mutex
	retriever *statistic.StatisticMetricRetriever
}

// NewCollector creates a new collector
func NewCollector(c *api.ApiClient) prometheus.Collector {
	r := statistic.NewStatisticMetricRetriever("vm", c)
	return &VmCollector{api: c, retriever: r}
}

// Collect implements Prometheus Collector interface
func (c *VmCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.getMetrics() {
		ch <- m
	}
}

// Describe implements Prometheus Collector interface
func (c *VmCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.getMetrics() {
		ch <- m.Desc()
	}
}

func (c *VmCollector) getMetrics() []prometheus.Metric {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.metrics != nil {
		return c.metrics
	}

	c.retrieveMetrics()
	return c.metrics
}

func (c *VmCollector) retrieveMetrics() {
	ressources := make(map[string]string)
	for _, vm := range c.getVms() {
		ressources[vm.Id] = vm.Name
	}

	c.metrics = c.retriever.RetrieveMetrics(ressources)
}

func (c *VmCollector) getVms() []Vm {
	var vms Vms
	err := c.api.GetAndParse("vms", &vms)

	if err != nil {
		log.Error(err)
	}

	return vms.Vm
}
