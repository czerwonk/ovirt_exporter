package storagedomain

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/metric"
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
	l := []string{"name", "type", "path"}
	availableDesc = prometheus.NewDesc(prefix+"available_bytes", "Available space in bytes", l, nil)
	usedDesc = prometheus.NewDesc(prefix+"used_bytes", "Used space in bytes", l, nil)
	commitedDesc = prometheus.NewDesc(prefix+"commited_bytes", "Commited space in bytes", l, nil)
	upDesc = prometheus.NewDesc(prefix+"up", "Status of storage domain", l, nil)
	masterDesc = prometheus.NewDesc(prefix+"master", "Storage domain is master", l, nil)
}

// StorageDomainCollector collects storage domain statistics from oVirt
type StorageDomainCollector struct {
	client *api.ApiClient
}

// NewCollector creates a new collector
func NewCollector(client *api.ApiClient) prometheus.Collector {
	return &StorageDomainCollector{client: client}
}

// Collect implements Prometheus Collector interface
func (c *StorageDomainCollector) Collect(ch chan<- prometheus.Metric) {
	s := StorageDomains{}
	err := c.client.GetAndParse("storagedomains", &s)
	if err != nil {
		log.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(s.Domains))

	for _, h := range s.Domains {
		go c.collectMetricsForDomain(h, ch, wg)
	}

	wg.Wait()
}

// Describe implements Prometheus Collector interface
func (c *StorageDomainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- masterDesc
	ch <- availableDesc
	ch <- usedDesc
	ch <- commitedDesc
}

func (c *StorageDomainCollector) collectMetricsForDomain(domain StorageDomain, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	d := &domain
	l := []string{d.Name, string(d.Type), d.Storage.Path}

	up := d.ExternalStatus == "ok"
	ch <- metric.MustCreate(upDesc, boolToFloat(up), l)
	ch <- metric.MustCreate(masterDesc, boolToFloat(d.Master), l)
	ch <- metric.MustCreate(availableDesc, float64(d.Available), l)
	ch <- metric.MustCreate(usedDesc, float64(d.Used), l)
	ch <- metric.MustCreate(commitedDesc, float64(d.Committed), l)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
