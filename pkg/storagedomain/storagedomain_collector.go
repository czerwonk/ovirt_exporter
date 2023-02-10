// SPDX-License-Identifier: MIT

package storagedomain

import (
	"github.com/czerwonk/ovirt_api/api"
	"github.com/czerwonk/ovirt_exporter/pkg/metric"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix = "ovirt_storage_"

var (
	availableDesc *prometheus.Desc
	usedDesc      *prometheus.Desc
	committedDesc *prometheus.Desc
	masterDesc    *prometheus.Desc
	upDesc        *prometheus.Desc
)

func init() {
	l := []string{"name", "type", "path"}
	availableDesc = prometheus.NewDesc(prefix+"available_bytes", "Available space in bytes", l, nil)
	usedDesc = prometheus.NewDesc(prefix+"used_bytes", "Used space in bytes", l, nil)
	committedDesc = prometheus.NewDesc(prefix+"committed_bytes", "Committed space in bytes", l, nil)
	upDesc = prometheus.NewDesc(prefix+"up", "Status of storage domain", l, nil)
	masterDesc = prometheus.NewDesc(prefix+"master", "Storage domain is master", l, nil)
}

// StorageDomainCollector collects storage domain statistics from oVirt
type StorageDomainCollector struct {
	client          *api.Client
	collectDuration prometheus.Observer
}

// NewCollector creates a new collector
func NewCollector(client *api.Client, collectDuration prometheus.Observer) prometheus.Collector {
	return &StorageDomainCollector{client: client, collectDuration: collectDuration}
}

// Collect implements Prometheus Collector interface
func (c *StorageDomainCollector) Collect(ch chan<- prometheus.Metric) {
	timer := prometheus.NewTimer(c.collectDuration)
	defer timer.ObserveDuration()

	s := StorageDomains{}
	err := c.client.GetAndParse("storagedomains", &s)
	if err != nil {
		log.Error(err)
		return
	}

	for _, h := range s.Domains {
		c.collectMetricsForDomain(h, ch)
	}
}

// Describe implements Prometheus Collector interface
func (c *StorageDomainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- masterDesc
	ch <- availableDesc
	ch <- usedDesc
	ch <- committedDesc
}

func (c *StorageDomainCollector) collectMetricsForDomain(domain StorageDomain, ch chan<- prometheus.Metric) {
	d := &domain
	l := []string{d.Name, string(d.Type), d.Storage.Path}

	up := d.ExternalStatus == "ok"
	ch <- metric.MustCreate(upDesc, boolToFloat(up), l)
	ch <- metric.MustCreate(masterDesc, boolToFloat(d.Master), l)
	ch <- metric.MustCreate(availableDesc, float64(d.Available), l)
	ch <- metric.MustCreate(usedDesc, float64(d.Used), l)
	ch <- metric.MustCreate(committedDesc, float64(d.Committed), l)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
