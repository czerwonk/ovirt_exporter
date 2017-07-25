package vm

import (
	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/cluster"
	"github.com/czerwonk/ovirt_exporter/host"
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
	api              *api.ApiClient
	metrics          []prometheus.Metric
	mutex            sync.Mutex
	retriever        *statistic.StatisticMetricRetriever
	hostRetriever    *host.HostRetriever
	clusterRetriever *cluster.ClusterRetriever
}

// NewCollector creates a new collector
func NewCollector(api *api.ApiClient) prometheus.Collector {
	r := statistic.NewStatisticMetricRetriever("vm", api, labelNames)
	h := host.NewRetriever(api)
	c := cluster.NewRetriever(api)
	return &VmCollector{api: api, retriever: r, hostRetriever: h, clusterRetriever: c}
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
	ids := make([]string, 0)
	labelValues := make(map[string][]string)

	c.metrics = make([]prometheus.Metric, 0)
	for _, vm := range c.getVms() {
		ids = append(ids, vm.Id)
		labelValues[vm.Id] = c.getLabelValues(&vm)

		c.addMetricsForVm(&vm, labelValues[vm.Id])
	}

	c.metrics = append(c.metrics, c.retriever.RetrieveMetrics(ids, labelValues)...)
}

func (c *VmCollector) addMetricsForVm(vm *Vm, labelValues []string) {
	c.metrics = append(c.metrics, c.upMetric(vm, labelValues))
	c.addMetric(cpuCoresDesc, float64(vm.Cpu.Topology.Cores), labelValues)
	c.addMetric(cpuThreadsDesc, float64(vm.Cpu.Topology.Threads), labelValues)
	c.addMetric(cpuSocketsDesc, float64(vm.Cpu.Topology.Sockets), labelValues)
}

func (c *VmCollector) addMetric(desc *prometheus.Desc, v float64, labelValues []string) {
	c.metrics = append(c.metrics, prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, labelValues...))
}

func (c *VmCollector) upMetric(vm *Vm, labelValues []string) prometheus.Metric {
	var up float64
	if vm.Status == "up" {
		up = 1
	}

	return prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, up, labelValues...)
}

func (c *VmCollector) getLabelValues(vm *Vm) []string {
	h := &host.Host{}
	var err error

	if len(vm.Host.Id) > 0 {
		h, err = c.hostRetriever.Get(vm.Host.Id)
		if err != nil {
			log.Error(err)
		}
	}

	cl, err := c.clusterRetriever.Get(vm.Cluster.Id)
	if err != nil {
		log.Error(err)
	}

	return []string{vm.Name, h.Name, cl.Name}
}

func (c *VmCollector) getVms() []Vm {
	var vms Vms
	err := c.api.GetAndParse("vms", &vms)

	if err != nil {
		log.Error(err)
	}

	return vms.Vm
}
