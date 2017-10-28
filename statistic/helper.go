package statistic

import (
	"strings"

	"github.com/czerwonk/ovirt_exporter/metric"
	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func CollectStatisticMetrics(prefix string, conn *ovirtsdk.Connection, stats *ovirtsdk.StatisticSlice, ch chan<- prometheus.Metric, labelNames []string, labelValues []string) {
	x, err := conn.FollowLink(stats)
	if err != nil {
		log.Error(err)
		return
	}

	stats = x.(*ovirtsdk.StatisticSlice)

	for _, s := range stats.Slice() {
		metricName := strings.Replace(s.MustName(), ".", "_", -1)

		if s.MustUnit() != "none" {
			metricName += "_" + string(s.MustUnit())
		}

		n := prefix + metricName
		d := prometheus.NewDesc(n, s.MustDescription(), labelNames, nil)

		ch <- metric.MustCreate(d, s.MustValues().Slice()[0].MustDatum(), labelValues)
	}
}
