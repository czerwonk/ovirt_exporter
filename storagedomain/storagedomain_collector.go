package storagedomain

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/imjoey/go-ovirt"
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
	conn *ovirtsdk4.Connection
}

// NewCollector creates a new collector
func NewCollector(conn *ovirtsdk4.Connection) prometheus.Collector {
	return &StorageDomainCollector{conn: conn}
}

// Collect implements Prometheus Collector interface
func (c *StorageDomainCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.conn.SystemService().StorageDomainsService()
	resp, err := s.List().Send()
	if err != nil {
		log.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	slice := resp.MustStorageDomains().Slice()
	wg.Add(len(slice))

	for _, h := range slice {
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

func (c *StorageDomainCollector) collectMetricsForDomain(domain ovirtsdk4.StorageDomain, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	d := &domain
	l := []string{d.MustName(), string(d.MustType()), c.storagePath(d)}

	up := d.MustExternalStatus() == "ok"
	ch <- metric.MustCreate(upDesc, boolToFloat(up), l)
	ch <- metric.MustCreate(masterDesc, boolToFloat(d.MustMaster()), l)

	if v, ok := d.Available(); ok {
		ch <- metric.MustCreate(availableDesc, float64(v), l)
	}

	if v, ok := d.Used(); ok {
		ch <- metric.MustCreate(usedDesc, float64(v), l)
	}

	if v, ok := d.Committed(); ok {
		ch <- metric.MustCreate(commitedDesc, float64(v), l)
	}
}

func (c *StorageDomainCollector) storagePath(d *ovirtsdk4.StorageDomain) string {
	s, ok := d.Storage()
	if !ok {
		return ""
	}

	p, ok := s.Path()
	if !ok {
		return ""
	}

	return p
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}

	return 0
}
