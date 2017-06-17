package vm

import (
	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/prometheus/client_golang/prometheus"
)

// VmCollector collects virtual machine statistics from oVirt
type VmCollector struct {
	api *api.ApiClient
}

// NeNewCollector creates a new collector
func NewCollector(c *api.ApiClient) prometheus.Collector {
	return &VmCollector{api: c}
}

// Collect implements Prometheus Collector interface
func (c *VmCollector) Collect(ch chan<- prometheus.Metric) {

}

// Describe implements Prometheus Collector interface
func (c *VmCollector) Describe(ch chan<- *prometheus.Desc) {

}
