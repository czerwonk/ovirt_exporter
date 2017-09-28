package vm

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/cluster"
	"github.com/czerwonk/ovirt_exporter/host"
	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/czerwonk/ovirt_exporter/statistic"
	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const prefix = "ovirt_vm_"

var (
	upDesc         *prometheus.Desc
	cpuCoresDesc   *prometheus.Desc
	cpuSocketsDesc *prometheus.Desc
	cpuThreadsDesc *prometheus.Desc
	labelNames     []string
)

func init() {
	labelNames = []string{"name", "host", "cluster"}
	upDesc = prometheus.NewDesc(prefix+"up", "VM is running (1) or not (0)", labelNames, nil)
	cpuCoresDesc = prometheus.NewDesc(prefix+"cpu_cores", "Number of CPU cores assigned", labelNames, nil)
	cpuSocketsDesc = prometheus.NewDesc(prefix+"cpu_sockets", "Number of sockets", labelNames, nil)
	cpuThreadsDesc = prometheus.NewDesc(prefix+"cpu_threads", "Number of threads", labelNames, nil)
}

// VmCollector collects virtual machine statistics from oVirt
type VmCollector struct {
	conn    *ovirtsdk4.Connection
	metrics []prometheus.Metric
	mutex   sync.Mutex
}

// NewCollector creates a new collector
func NewCollector(conn *ovirtsdk4.Connection) prometheus.Collector {
	return &VmCollector{conn: conn}
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
	s := c.conn.SystemService().VmsService()
	resp, err := s.List().Send()
	if err != nil {
		log.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	slice := resp.MustVms().Slice()
	wg.Add(len(slice))

	ch := make(chan prometheus.Metric)
	for _, v := range slice {
		go c.collectForVm(v, ch, wg)
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

func (c *VmCollector) collectForVm(vm ovirtsdk4.Vm, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	v := &vm
	l := []string{v.MustName(), c.hostName(v), c.clusterName(v)}

	ch <- c.upMetric(v, l)

	c.collectCpuMetrics(v, ch, l)

	if stats, ok := v.Statistics(); ok {
		statistic.CollectStatisticMetrics(prefix, c.conn, stats, ch, labelNames, l)
	}
}

func (c *VmCollector) collectCpuMetrics(vm *ovirtsdk4.Vm, ch chan<- prometheus.Metric, l []string) {
	topo := vm.MustCpu().MustTopology()
	ch <- metric.MustCreate(cpuCoresDesc, float64(topo.MustCores()), l)
	ch <- metric.MustCreate(cpuThreadsDesc, float64(topo.MustThreads()), l)
	ch <- metric.MustCreate(cpuSocketsDesc, float64(topo.MustSockets()), l)
}

func (c *VmCollector) hostName(vm *ovirtsdk4.Vm) string {
	h, ok := vm.Host()
	if !ok {
		return ""
	}

	h, err := host.Follow(h, c.conn)
	if err != nil {
		log.Error(err)
		return ""
	}

	return h.MustName()
}

func (c *VmCollector) clusterName(vm *ovirtsdk4.Vm) string {
	cl, err := cluster.Follow(vm.MustCluster(), c.conn)
	if err != nil {
		log.Error(err)
		return ""
	}

	return cl.MustName()
}

func (c *VmCollector) upMetric(vm *ovirtsdk4.Vm, labelValues []string) prometheus.Metric {
	var up float64
	if vm.MustStatus() == "up" {
		up = 1
	}

	return metric.MustCreate(upDesc, up, labelValues)
}
