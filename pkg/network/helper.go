// SPDX-License-Identifier: MIT

package network

import (
	"fmt"
	"sync"

	"github.com/czerwonk/ovirt_exporter/pkg/collector.go"
	"github.com/czerwonk/ovirt_exporter/pkg/statistic"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

// CollectMetricsForHost collects net metrics for a specific Host
func CollectMetricsForHost(ctx context.Context, path, prefix string, labelNames, labelValues []string, cc *collector.CollectorContext) error {
	ctx, span := cc.Tracer().Start(ctx, "Network.CollectForHost", trace.WithAttributes(
		attribute.String("prefix", prefix),
	))
	defer span.End()

	nics := &HostNics{}
	err := cc.Client().GetAndParse(ctx, path, nics)
	if err != nil {
		return err
	}

	return collectForNICs(ctx, nics.Nics, path, prefix, labelNames, labelValues, cc)
}

// CollectMetricsForVM collects net metrics for a specific VM
func CollectMetricsForVM(ctx context.Context, path, prefix string, labelNames, labelValues []string, cc *collector.CollectorContext) error {
	ctx, span := cc.Tracer().Start(ctx, "Network.CollectForVM", trace.WithAttributes(
		attribute.String("prefix", prefix),
	))
	defer span.End()

	nics := &VMNics{}
	err := cc.Client().GetAndParse(ctx, path, nics)
	if err != nil {
		return err
	}

	return collectForNICs(ctx, nics.Nics, path, prefix, labelNames, labelValues, cc)
}

func collectForNICs(ctx context.Context, nics []Nic, path, prefix string, labelNames, labelValues []string, cc *collector.CollectorContext) error {
	wg := sync.WaitGroup{}
	wg.Add(len(nics))
	for _, n := range nics {
		p := fmt.Sprintf("%s/%s/statistics", path, n.ID)
		ln := append(labelNames, "nic", "mac")
		l := append(labelValues, n.Name, n.Mac.Address)

		go func() {
			statistic.CollectMetrics(ctx, p, prefix+"network_", ln, l, cc)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil
}
