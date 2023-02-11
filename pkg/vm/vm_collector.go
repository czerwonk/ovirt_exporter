// SPDX-License-Identifier: MIT

package vm

import (
	"context"
	"sync"

	"fmt"

	"time"

	"github.com/czerwonk/ovirt_exporter/pkg/cluster"
	"github.com/czerwonk/ovirt_exporter/pkg/collector.go"
	"github.com/czerwonk/ovirt_exporter/pkg/disk"
	"github.com/czerwonk/ovirt_exporter/pkg/host"
	"github.com/czerwonk/ovirt_exporter/pkg/metric"
	"github.com/czerwonk/ovirt_exporter/pkg/network"
	"github.com/czerwonk/ovirt_exporter/pkg/statistic"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	cc               *collector.CollectorContext
	collectDuration  prometheus.Observer
	metrics          []prometheus.Metric
	collectSnapshots bool
	collectNetwork   bool
	collectDisks     bool
	mutex            sync.Mutex
	rootCtx          context.Context
}

// NewCollector creates a new collector
func NewCollector(ctx context.Context, cc *collector.CollectorContext, collectSnaphots, collectNetwork bool, collectDisks bool, collectDuration prometheus.Observer) prometheus.Collector {
	return &VMCollector{
		cc:               cc,
		collectSnapshots: collectSnaphots,
		collectNetwork:   collectNetwork,
		collectDisks:     collectDisks,
		collectDuration:  collectDuration,
		rootCtx:          ctx,
	}
}

// Collect implements Prometheus Collector interface
func (c *VMCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, span := c.cc.Tracer().Start(c.rootCtx, "VMCollector.Collect")
	defer span.End()

	for _, m := range c.getMetrics(ctx, span) {
		ch <- m
	}
}

// Describe implements Prometheus Collector interface
func (c *VMCollector) Describe(ch chan<- *prometheus.Desc) {
	ctx, span := c.cc.Tracer().Start(c.rootCtx, "VMCollector.Describe")
	defer span.End()

	for _, m := range c.getMetrics(ctx, span) {
		ch <- m.Desc()
	}
}

func (c *VMCollector) getMetrics(ctx context.Context, span trace.Span) []prometheus.Metric {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.metrics != nil {
		return c.metrics
	}

	c.retrieveMetrics(ctx, span)
	return c.metrics
}

func (c *VMCollector) retrieveMetrics(ctx context.Context, span trace.Span) {
	timer := prometheus.NewTimer(c.collectDuration)
	defer timer.ObserveDuration()

	v := VMs{}
	err := c.cc.Client().GetAndParse(ctx, "vms", &v)
	if err != nil {
		c.cc.HandleError(err, span)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(v.VMs))

	ch := make(chan prometheus.Metric)
	c.cc.SetMetricsCh(ch)
	for _, v := range v.VMs {
		go c.collectForVM(ctx, v, wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for m := range ch {
		c.metrics = append(c.metrics, m)
	}
}

func (c *VMCollector) collectForVM(ctx context.Context, vm VM, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, span := c.cc.Tracer().Start(ctx, "VMCollector.CollectForVM", trace.WithAttributes(
		attribute.String("vm_name", vm.Name),
		attribute.String("vm_id", vm.ID),
	))
	defer span.End()

	v := &vm
	l := []string{v.Name, c.hostName(ctx, v), cluster.Name(ctx, v.Cluster.ID, c.cc.Client())}

	c.cc.RecordMetrics(
		c.upMetric(v, l),
		c.diskImageIllegalMetric(v, l),
	)

	c.collectCPUMetrics(v, l)

	statPath := fmt.Sprintf("vms/%s/statistics", vm.ID)
	statistic.CollectMetrics(ctx, statPath, prefix, labelNames, l, c.cc)

	if c.collectNetwork {
		networkPath := fmt.Sprintf("vms/%s/nics", vm.ID)
		network.CollectMetricsForVM(ctx, networkPath, prefix, labelNames, l, c.cc)
	}

	if c.collectSnapshots {
		c.collectSnapshotMetrics(ctx, v, l)
	}

	if c.collectDisks {
		c.collectDiskMetrics(ctx, v, l)
	}
}

func (c *VMCollector) collectCPUMetrics(vm *VM, l []string) {
	topo := vm.CPU.Topology

	c.cc.RecordMetrics(
		metric.MustCreate(cpuCoresDesc, float64(topo.Cores), l),
		metric.MustCreate(cpuThreadsDesc, float64(topo.Threads), l),
		metric.MustCreate(cpuSocketsDesc, float64(topo.Sockets), l),
	)
}

func (c *VMCollector) hostName(ctx context.Context, vm *VM) string {
	if len(vm.Host.ID) == 0 {
		return ""
	}

	return host.Name(ctx, vm.Host.ID, c.cc.Client())
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

func (c *VMCollector) collectSnapshotMetrics(ctx context.Context, vm *VM, l []string) {
	ctx, span := c.cc.Tracer().Start(ctx, "VMCollector.CollectSnapshotMetrics")
	defer span.End()

	snaps := Snapshots{}
	path := fmt.Sprintf("vms/%s/snapshots", vm.ID)

	err := c.cc.Client().GetAndParse(ctx, path, &snaps)
	if err != nil {
		c.cc.HandleError(err, span)
		return
	}

	s := snaps.Snapshot[1:]
	c.cc.RecordMetrics(
		metric.MustCreate(snapshotCount, float64(len(s)), l),
	)

	if len(s) == 0 {
		return
	}

	min := s[0]
	max := s[len(s)-1]
	c.cc.RecordMetrics(
		metric.MustCreate(maxSnapshotAge, time.Since(min.Date).Seconds(), l),
		metric.MustCreate(minSnapshotAge, time.Since(max.Date).Seconds(), l),
	)
}

func (c *VMCollector) collectDiskMetrics(ctx context.Context, vm *VM, l []string) {
	ctx, span := c.cc.Tracer().Start(ctx, "VMCollector.CollectDiskMetrics")
	defer span.End()

	attchs := DiskAttachments{}
	path := fmt.Sprintf("vms/%s/diskattachments", vm.ID)

	err := c.cc.Client().GetAndParse(ctx, path, &attchs)
	if err != nil {
		c.cc.HandleError(err, span)
		return
	}

	if len(attchs.Attachment) == 0 {
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(attchs.Attachment))

	for _, a := range attchs.Attachment {
		go c.collectForAttachment(ctx, a, l, wg)
	}

	wg.Wait()
}

func (c *VMCollector) collectForAttachment(ctx context.Context, attachment DiskAttachment, l []string, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, span := c.cc.Tracer().Start(ctx, "VMCollector.CollectAttachement")
	defer span.End()

	d, err := disk.Get(ctx, attachment.Disk.ID, c.cc.Client())
	if err != nil {
		c.cc.HandleError(err, span)
		return
	}

	if d == nil {
		c.cc.HandleError(fmt.Errorf("could not find disk with ID %s", attachment.Disk.ID), span)
		return
	}

	l = append(l, d.Name, d.Alias, attachment.LogicalName, d.StorageDomainName())

	c.cc.RecordMetrics(
		metric.MustCreate(diskProvisionedSize, float64(d.ProvisionedSize), l),
		metric.MustCreate(diskActualSize, float64(d.ActualSize), l),
		metric.MustCreate(diskTotalSize, float64(d.TotalSize), l),
	)
}
