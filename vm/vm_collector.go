// SPDX-License-Identifier: MIT

package vm

import (
	"sync"

	"fmt"

	"time"

	"github.com/czerwonk/ovirt_api/api"
	"github.com/czerwonk/ovirt_exporter/cluster"
	"github.com/czerwonk/ovirt_exporter/disk"
	"github.com/czerwonk/ovirt_exporter/host"
	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/czerwonk/ovirt_exporter/network"
	"github.com/czerwonk/ovirt_exporter/statistic"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix = "ovirt_vm_"

var (
	upDesc              *prometheus.Desc
	cpuCoresDesc        *prometheus.Desc
	cpuSocketsDesc      *prometheus.Desc
	cpuThreadsDesc      *prometheus.Desc
	snapshotCount       *prometheus.Desc
	minSnapshotAge      *prometheus.Desc
	maxSnapshotAge      *prometheus.Desc
	illegalImages       *prometheus.Desc
	diskProvisionedSize *prometheus.Desc
	diskActualSize      *prometheus.Desc
	diskTotalSize       *prometheus.Desc
	labelNames          []string
)

func init() {
	labelNames = []string{"name", "host", "cluster"}
	upDesc = prometheus.NewDesc(prefix+"up", "VM is running (1) or not (0)", labelNames, nil)
	cpuCoresDesc = prometheus.NewDesc(prefix+"cpu_cores", "Number of CPU cores assigned", labelNames, nil)
	cpuSocketsDesc = prometheus.NewDesc(prefix+"cpu_sockets", "Number of sockets", labelNames, nil)
	cpuThreadsDesc = prometheus.NewDesc(prefix+"cpu_threads", "Number of threads", labelNames, nil)
	snapshotCount = prometheus.NewDesc(prefix+"snapshots", "Number of snapshots", labelNames, nil)
	maxSnapshotAge = prometheus.NewDesc(prefix+"snapshot_max_age_seconds", "Age of the oldest snapshot in seconds", labelNames, nil)
	minSnapshotAge = prometheus.NewDesc(prefix+"snapshot_min_age_seconds", "Age of the newest snapshot in seconds", labelNames, nil)
	illegalImages = prometheus.NewDesc(prefix+"illegal_images", "Health status of the disks attatched to the VM (1 if one or more disk is in illegal state)", labelNames, nil)

	diskLabelNames := append(labelNames, "disk_name", "disk_alias", "disk_logical_name", "storage_domain")
	diskProvisionedSize = prometheus.NewDesc(prefix+"disk_provisioned_size_bytes", "Provisioned size of the disk in bytes", diskLabelNames, nil)
	diskActualSize = prometheus.NewDesc(prefix+"disk_actual_size_bytes", "Actual size of the disk in bytes", diskLabelNames, nil)
	diskTotalSize = prometheus.NewDesc(prefix+"disk_total_size_bytes", "Total size of the disk in bytes", diskLabelNames, nil)

}

// VMCollector collects virtual machine statistics from oVirt
type VMCollector struct {
	client           *api.Client
	collectDuration  prometheus.Observer
	metrics          []prometheus.Metric
	collectSnapshots bool
	collectNetwork   bool
	collectDisks     bool
	mutex            sync.Mutex
}

// NewCollector creates a new collector
func NewCollector(client *api.Client, collectSnaphots, collectNetwork bool, collectDisks bool, collectDuration prometheus.Observer) prometheus.Collector {
	return &VMCollector{
		client:           client,
		collectSnapshots: collectSnaphots,
		collectNetwork:   collectNetwork,
		collectDisks:     collectDisks,
		collectDuration:  collectDuration}
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
	timer := prometheus.NewTimer(c.collectDuration)
	defer timer.ObserveDuration()

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

	go func() {
		wg.Wait()
		close(ch)
	}()

	for m := range ch {
		c.metrics = append(c.metrics, m)
	}
}

func (c *VMCollector) collectForVM(vm VM, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	v := &vm
	l := []string{v.Name, c.hostName(v), cluster.Name(v.Cluster.ID, c.client)}

	ch <- c.upMetric(v, l)
	ch <- c.diskImageIllegalMetric(v, l)

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

	if c.collectDisks {
		c.collectDiskMetrics(v, ch, l)
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

func (c *VMCollector) diskImageIllegalMetric(vm *VM, labelValues []string) prometheus.Metric {
	var hasIllegalImages float64
	if vm.HasIllegalImages {
		hasIllegalImages = 1
	}

	return metric.MustCreate(illegalImages, hasIllegalImages, labelValues)
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

func (c *VMCollector) collectDiskMetrics(vm *VM, ch chan<- prometheus.Metric, l []string) {
	attchs := DiskAttachments{}
	path := fmt.Sprintf("vms/%s/diskattachments", vm.ID)

	err := c.client.GetAndParse(path, &attchs)
	if err != nil {
		log.Error(err)
		return
	}

	if len(attchs.Attachment) == 0 {
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(attchs.Attachment))

	for _, a := range attchs.Attachment {
		go c.collectForAttachment(a, ch, l, wg)
	}

	wg.Wait()
}

func (c *VMCollector) collectForAttachment(attachment DiskAttachment, ch chan<- prometheus.Metric, l []string, wg *sync.WaitGroup) {
	defer wg.Done()

	d, err := disk.Get(attachment.Disk.ID, c.client)
	if err != nil {
		log.Error(err)
		return
	}

	if d == nil {
		log.Error("could not find disk with ID " + attachment.Disk.ID)
		return
	}

	l = append(l, d.Name, d.Alias, attachment.LogicalName, d.StorageDomainName())
	ch <- metric.MustCreate(diskProvisionedSize, float64(d.ProvisionedSize), l)
	ch <- metric.MustCreate(diskActualSize, float64(d.ActualSize), l)
	ch <- metric.MustCreate(diskTotalSize, float64(d.TotalSize), l)
}
