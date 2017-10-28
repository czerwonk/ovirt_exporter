package host

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/cluster"
	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/czerwonk/ovirt_exporter/statistic"
	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "ovirt_host_"

var (
	upDesc         *prometheus.Desc
	cpuCoresDesc   *prometheus.Desc
	cpuSocketsDesc *prometheus.Desc
	cpuThreadsDesc *prometheus.Desc
	cpuSpeedDesc   *prometheus.Desc
	memoryDesc     *prometheus.Desc
	labelNames     []string
)

func init() {
	labelNames = []string{"name", "cluster"}
	upDesc = prometheus.NewDesc(prefix+"up", "Host is running (1) or not (0)", labelNames, nil)
	cpuCoresDesc = prometheus.NewDesc(prefix+"cpu_cores", "Number of CPU cores assigned", labelNames, nil)
	cpuSocketsDesc = prometheus.NewDesc(prefix+"cpu_sockets", "Number of sockets", labelNames, nil)
	cpuThreadsDesc = prometheus.NewDesc(prefix+"cpu_threads", "Number of threads", labelNames, nil)
	cpuSpeedDesc = prometheus.NewDesc(prefix+"cpu_speed", "CPU speed in MHz", labelNames, nil)
	memoryDesc = prometheus.NewDesc(prefix+"memory_installed_bytes", "Memory installed in bytes", labelNames, nil)
}

// HostCollector collects host statistics from oVirt
type HostCollector struct {
	conn    *ovirtsdk.Connection
	metrics []prometheus.Metric
	mutex   sync.Mutex
}

// NewCollector creates a new collector
func NewCollector(conn *ovirtsdk.Connection) prometheus.Collector {
	return &HostCollector{conn: conn}
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
	s := c.conn.SystemService().HostsService()
	resp, err := s.List().Send()
	if err != nil {
		log.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	slice := resp.MustHosts().Slice()
	wg.Add(len(slice))

	ch := make(chan prometheus.Metric)
	for _, h := range slice {
		go c.collectForHost(h, ch, wg)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	for {
		select {
		case m := <-ch:
			c.metrics = append(c.metrics, m)

		case <-done:
			return
		}
	}
}

func (c *HostCollector) collectForHost(host ovirtsdk.Host, ch chan prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	h := &host
	l := []string{h.MustName(), cluster.Name(h.MustCluster(), c.conn)}

	ch <- c.upMetric(h, l)
	ch <- metric.MustCreate(memoryDesc, float64(host.MustMemory()), l)
	c.collectCpuMetrics(h, ch, l)

	if stats, ok := h.Statistics(); ok {
		statistic.CollectStatisticMetrics(prefix, c.conn, stats, ch, labelNames, l)
	}
}

func (c *HostCollector) collectCpuMetrics(host *ovirtsdk.Host, ch chan prometheus.Metric, l []string) {
	topo := host.MustCpu().MustTopology()
	ch <- metric.MustCreate(cpuCoresDesc, float64(topo.MustCores()), l)
	ch <- metric.MustCreate(cpuThreadsDesc, float64(topo.MustThreads()), l)
	ch <- metric.MustCreate(cpuSocketsDesc, float64(topo.MustSockets()), l)
	ch <- metric.MustCreate(cpuSpeedDesc, float64(host.MustCpu().MustSpeed()), l)
}

func (c *HostCollector) addMetric(desc *prometheus.Desc, v float64, labelValues []string) {
	c.metrics = append(c.metrics, prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, labelValues...))
}

func (c *HostCollector) upMetric(h *ovirtsdk.Host, labelValues []string) prometheus.Metric {
	var up float64
	if h.MustStatus() == "up" {
		up = 1
	}

	return metric.MustCreate(upDesc, up, labelValues)
}
