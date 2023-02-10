// SPDX-License-Identifier: MIT

package network

import (
	"fmt"
	"sync"

	"github.com/czerwonk/ovirt_exporter/pkg/client"
	"github.com/czerwonk/ovirt_exporter/pkg/statistic"

	"github.com/prometheus/client_golang/prometheus"
)

// CollectMetricsForHost collects net metrics for a specific Host
func CollectMetricsForHost(path, prefix string, labelNames, labelValues []string, cl client.Client, ch chan<- prometheus.Metric) error {
	nics := &HostNics{}
	err := cl.GetAndParse(path, nics)
	if err != nil {
		return err
	}

	return collectForNics(nics.Nics, path, prefix, labelNames, labelValues, cl, ch)
}

// CollectMetricsForVM collects net metrics for a specific VM
func CollectMetricsForVM(path, prefix string, labelNames, labelValues []string, cl client.Client, ch chan<- prometheus.Metric) error {
	nics := &VMNics{}
	err := cl.GetAndParse(path, nics)
	if err != nil {
		return err
	}

	return collectForNics(nics.Nics, path, prefix, labelNames, labelValues, cl, ch)
}

func collectForNics(nics []Nic, path, prefix string, labelNames, labelValues []string, cl client.Client, ch chan<- prometheus.Metric) error {
	wg := sync.WaitGroup{}
	wg.Add(len(nics))
	for _, n := range nics {
		p := fmt.Sprintf("%s/%s/statistics", path, n.ID)
		ln := append(labelNames, "nic", "mac")
		l := append(labelValues, n.Name, n.Mac.Address)

		go func() {
			statistic.CollectMetrics(p, prefix+"network_", ln, l, cl, ch)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}
