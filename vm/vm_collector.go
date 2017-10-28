package vm

import (
	"sync"

	"time"

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
	snapshotCount  *prometheus.Desc
	minSnapshotAge *prometheus.Desc
	maxSnapshotAge *prometheus.Desc
	labelNames     []string
)

func init() {
	labelNames = []string{"name", "host", "cluster"}
	upDesc = prometheus.NewDesc(prefix+"up", "VM is running (1) or not (0)", labelNames, nil)
	cpuCoresDesc = prometheus.NewDesc(prefix+"cpu_cores", "Number of CPU cores assigned", labelNames, nil)
	cpuSocketsDesc = prometheus.NewDesc(prefix+"cpu_sockets", "Number of sockets", labelNames, nil)
	cpuThreadsDesc = prometheus.NewDesc(prefix+"cpu_threads", "Number of threads", labelNames, nil)
	snapshotCount = prometheus.NewDesc(prefix+"snapshot_count", "Number of snapshots", labelNames, nil)
	maxSnapshotAge = prometheus.NewDesc(prefix+"snapshot_max_age_minutes", "Age of the oldest snapshot in minutes", labelNames, nil)
	minSnapshotAge = prometheus.NewDesc(prefix+"snapshot_min_age_minutes", "Age of the newest snapshot in minutes", labelNames, nil)
}

// VmCollector collects virtual machine statistics from oVirt
type VmCollector struct {
	conn             *ovirtsdk.Connection
	metrics          []prometheus.Metric
	collectSnapshots bool
	mutex            sync.Mutex
}

// NewCollector creates a new collector
func NewCollector(conn *ovirtsdk.Connection, collectSnaphots bool) prometheus.Collector {
	return &VmCollector{conn: conn, collectSnapshots: collectSnaphots}
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

func (c *VmCollector) collectForVm(vm ovirtsdk.Vm, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	v := &vm
	l := []string{v.MustName(), c.hostName(v), cluster.Name(v.MustCluster(), c.conn)}

	ch <- c.upMetric(v, l)

	c.collectCpuMetrics(v, ch, l)

	if stats, ok := v.Statistics(); ok {
		statistic.CollectStatisticMetrics(prefix, c.conn, stats, ch, labelNames, l)
	}

	if c.collectSnapshots {
		c.collectSnapshotMetrics(v, ch, l)
	}
}

func (c *VmCollector) collectCpuMetrics(vm *ovirtsdk.Vm, ch chan<- prometheus.Metric, l []string) {
	topo := vm.MustCpu().MustTopology()
	ch <- metric.MustCreate(cpuCoresDesc, float64(topo.MustCores()), l)
	ch <- metric.MustCreate(cpuThreadsDesc, float64(topo.MustThreads()), l)
	ch <- metric.MustCreate(cpuSocketsDesc, float64(topo.MustSockets()), l)
}

func (c *VmCollector) hostName(vm *ovirtsdk.Vm) string {
	h, ok := vm.Host()
	if !ok {
		return ""
	}

	return host.Name(h, c.conn)
}

func (c *VmCollector) upMetric(vm *ovirtsdk.Vm, labelValues []string) prometheus.Metric {
	var up float64
	if vm.MustStatus() == "up" {
		up = 1
	}

	return metric.MustCreate(upDesc, up, labelValues)
}

func (c *VmCollector) collectSnapshotMetrics(vm *ovirtsdk.Vm, ch chan<- prometheus.Metric, l []string) {
	snaps, err := followSnapShots(vm.MustSnapshots(), c.conn)
	if err != nil {
		log.Error(err)
		return
	}

	s := snaps.Slice()[1:]
	ch <- metric.MustCreate(snapshotCount, float64(len(s)), l)

	if len(s) == 0 {
		return
	}

	min := s[0]
	ch <- metric.MustCreate(maxSnapshotAge, time.Since(min.MustDate()).Seconds(), l)

	max := s[len(s)-1]
	ch <- metric.MustCreate(minSnapshotAge, time.Since(max.MustDate()).Seconds(), l)
}
