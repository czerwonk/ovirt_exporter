package host

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/cluster"
	"github.com/czerwonk/ovirt_exporter/statistic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "ovirt_host_"

var (
	upDesc     *prometheus.Desc
	labelNames []string
)

func init() {
	labelNames = []string{"name", "cluster"}
	upDesc = prometheus.NewDesc(prefix+"up", "Host is running (1) or not (0)", labelNames, nil)
}

// HostCollector collects host statistics from oVirt
type HostCollector struct {
	api              *api.ApiClient
	metrics          []prometheus.Metric
	mutex            sync.Mutex
	retriever        *statistic.StatisticMetricRetriever
	clusterRetriever *cluster.ClusterRetriever
}

// NewCollector creates a new collector
func NewCollector(api *api.ApiClient) prometheus.Collector {
	r := statistic.NewStatisticMetricRetriever("host", api, labelNames)
	c := cluster.NewRetriever(api)
	return &HostCollector{api: api, retriever: r, clusterRetriever: c}
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
	ids := make([]string, 0)
	labelValues := make(map[string][]string)

	c.metrics = make([]prometheus.Metric, 0)
	for _, h := range c.getHosts() {
		ids = append(ids, h.Id)
		cluster, err := c.clusterRetriever.Get(h.Cluster.Id)
		if err != nil {
			log.Error(err)
		}

		labelValues[h.Id] = []string{h.Name, cluster.Name}

		c.metrics = append(c.metrics, c.upMetric(&h, labelValues[h.Id]))
	}

	c.metrics = append(c.metrics, c.retriever.RetrieveMetrics(ids, labelValues)...)
}

func (c *HostCollector) upMetric(h *Host, labelValues []string) prometheus.Metric {
	var up float64
	if h.Status == "up" {
		up = 1
	}

	return prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, up, labelValues...)
}

func (c *HostCollector) getHosts() []Host {
	var hosts Hosts
	err := c.api.GetAndParse("hosts", &hosts)

	if err != nil {
		log.Error(err)
	}

	return hosts.Host
}
