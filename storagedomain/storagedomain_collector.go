package storagedomain

import (
	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/datacenter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "ovirt_storage_"

var (
	availableDesc *prometheus.Desc
	usedDesc      *prometheus.Desc
	commitedDesc  *prometheus.Desc
	masterDesc    *prometheus.Desc
	upDesc        *prometheus.Desc
)

func init() {
	l := []string{"name", "type", "path", "datacenter"}
	availableDesc = prometheus.NewDesc(prefix+"available_bytes", "Available space in bytes", l, nil)
	usedDesc = prometheus.NewDesc(prefix+"used_bytes", "Used space in bytes", l, nil)
	commitedDesc = prometheus.NewDesc(prefix+"commited_bytes", "Commited space in bytes", l, nil)
	upDesc = prometheus.NewDesc(prefix+"up", "Status of storage domain", l, nil)
	masterDesc = prometheus.NewDesc(prefix+"master", "Storage domain is master", l, nil)
}

// StorageDomainCollector collects storage domain statistics from oVirt
type StorageDomainCollector struct {
	api                 *api.ApiClient
	datacenterRetriever *datacenter.DatacenterRetriever
}

// NewCollector creates a new collector
func NewCollector(c *api.ApiClient) prometheus.Collector {
	dc := datacenter.NewRetriever(c)
	return &StorageDomainCollector{api: c, datacenterRetriever: dc}
}

// Collect implements Prometheus Collector interface
func (c *StorageDomainCollector) Collect(ch chan<- prometheus.Metric) {
	for _, d := range c.getDomains() {
		c.collectMetricsForDomain(&d, ch)
	}
}

// Describe implements Prometheus Collector interface
func (c *StorageDomainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- masterDesc
	ch <- availableDesc
	ch <- usedDesc
	ch <- commitedDesc
}

func (c *StorageDomainCollector) getDomains() []StorageDomain {
	var domains StorageDomains
	err := c.api.GetAndParse("storagedomains", &domains)

	if err != nil {
		log.Error(err)
	}

	return domains.Domains
}

func (c *StorageDomainCollector) collectMetricsForDomain(d *StorageDomain, ch chan<- prometheus.Metric) {
	dc, err := c.datacenterRetriever.Get(d.DataCenters.DataCenter.Id)
	if err != nil {
		log.Error(err)
	}

	l := []string{d.Name, d.Type, d.Storage.Path, dc.Name}

	up := d.ExternalStatus == "ok"
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, boolToFloat(up), l...)
	ch <- prometheus.MustNewConstMetric(masterDesc, prometheus.GaugeValue, boolToFloat(d.Master), l...)
	ch <- prometheus.MustNewConstMetric(availableDesc, prometheus.GaugeValue, d.Available, l...)
	ch <- prometheus.MustNewConstMetric(usedDesc, prometheus.GaugeValue, d.Used, l...)
	ch <- prometheus.MustNewConstMetric(commitedDesc, prometheus.GaugeValue, d.Committed, l...)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
