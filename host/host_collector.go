package host

import (
	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/prometheus/client_golang/prometheus"
)

// HostCollector collects virtual machine statistics from oVirt
type HostCollector struct {
	api *api.ApiClient
}

// NeNewCollector creates a new collector
func NewCollector(c *api.ApiClient) prometheus.Collector {
	return &HostCollector{api: c}
}

// Collect implements Prometheus Collector interface
func (c *HostCollector) Collect(ch chan<- prometheus.Metric) {

}

// Describe implements Prometheus Collector interface
func (c *HostCollector) Describe(ch chan<- *prometheus.Desc) {

}
