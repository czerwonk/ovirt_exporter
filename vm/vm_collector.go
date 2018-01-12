package vm

import (
	"sync"

	"fmt"

	"time"

	"github.com/czerwonk/ovirt_api/api"
	"github.com/czerwonk/ovirt_exporter/cluster"
	"github.com/czerwonk/ovirt_exporter/host"
	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/czerwonk/ovirt_exporter/network"
	"github.com/czerwonk/ovirt_exporter/statistic"
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

// VMCollector collects virtual machine statistics from oVirt
type VMCollector struct {
	client           *api.Client
	metrics          []prometheus.Metric
	collectSnapshots bool
	collectNetwork   bool
	mutex            sync.Mutex
}

// NewCollector creates a new collector
func NewCollector(client *api.Client, collectSnaphots, collectNetwork bool) prometheus.Collector {
	return &VMCollector{client: client, collectSnapshots: collectSnaphots, collectNetwork: collectNetwork}
}

// Collect implements Prometheus Collector interface
func (c *VMCollector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.getMetrics() {
		ch <- m
	}
}

// Describe implements Prometheus Collector interface
func (c *VMCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.getMetrics() {
		ch <- m.Desc()
	}
}

func (c *VMCollector) getMetrics() []prometheus.Metric {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.metrics != nil {
		return c.metrics
	}

	c.retrieveMetrics()
	return c.metrics
}

func (c *VMCollector) retrieveMetrics() {
	v := VMs{}
	err := c.client.GetAndParse("vms", &v)
	if err != nil {
		log.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(v.VMs))

	ch := make(chan prometheus.Metric)
	for _, v := range v.VMs {
		go c.collectForVM(v, ch, wg)
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

func (c *VMCollector) collectForVM(vm VM, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	v := &vm
	l := []string{v.Name, c.hostName(v), cluster.Name(v.Cluster.ID, c.client)}

	ch <- c.upMetric(v, l)

	c.collectCPUMetrics(v, ch, l)

	statPath := fmt.Sprintf("vms/%s/statistics", vm.ID)
	statistic.CollectMetrics(statPath, prefix, labelNames, l, c.client, ch)

	if c.collectNetwork {
		networkPath := fmt.Sprintf("vms/%s/nics", vm.ID)
		network.CollectMetricsForVM(networkPath, prefix, labelNames, l, c.client, ch)
	}

	if c.collectSnapshots {
		c.collectSnapshotMetrics(v, ch, l)
	}
}

func (c *VMCollector) collectCPUMetrics(vm *VM, ch chan<- prometheus.Metric, l []string) {
	topo := vm.CPU.Topology
	ch <- metric.MustCreate(cpuCoresDesc, float64(topo.Cores), l)
	ch <- metric.MustCreate(cpuThreadsDesc, float64(topo.Threads), l)
	ch <- metric.MustCreate(cpuSocketsDesc, float64(topo.Sockets), l)
}

func (c *VMCollector) hostName(vm *VM) string {
	if len(vm.Host.ID) == 0 {
		return ""
	}

	return host.Name(vm.Host.ID, c.client)
}

func (c *VMCollector) upMetric(vm *VM, labelValues []string) prometheus.Metric {
	var up float64
	if vm.Status == "up" {
		up = 1
	}

	return metric.MustCreate(upDesc, up, labelValues)
}

func (c *VMCollector) collectSnapshotMetrics(vm *VM, ch chan<- prometheus.Metric, l []string) {
	snaps := Snapshots{}
	path := fmt.Sprintf("vms/%s/snapshots", vm.ID)

	err := c.client.GetAndParse(path, &snaps)
	if err != nil {
		log.Error(err)
		return
	}

	s := snaps.Snapshot[1:]
	ch <- metric.MustCreate(snapshotCount, float64(len(s)), l)

	if len(s) == 0 {
		return
	}

	min := s[0]
	ch <- metric.MustCreate(maxSnapshotAge, time.Since(min.Date).Seconds(), l)

	max := s[len(s)-1]
	ch <- metric.MustCreate(minSnapshotAge, time.Since(max.Date).Seconds(), l)
}
