package statistic

import (
	"sync"

	"strings"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type Statistics struct {
	Statistic []Statistic `xml:"statistic"`
}

type Statistic struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Kind        string `xml:"kind"`
	Type        string `xml:"type"`
	Unit        string `xml:"unit"`
	Values      struct {
		Value struct {
			Datum float64 `xml:"datum"`
		} `xml:"value"`
	} `xml:"values"`
}

type StatisticMetricRetriever struct {
	ressource  string
	api        *api.ApiClient
	labelNames []string
}

func NewStatisticMetricRetriever(ressource string, api *api.ApiClient, labelNames []string) *StatisticMetricRetriever {
	return &StatisticMetricRetriever{ressource: ressource, api: api, labelNames: labelNames}
}

func (m *StatisticMetricRetriever) RetrieveMetrics(ids []string, labelValues map[string][]string) []prometheus.Metric {
	wg := &sync.WaitGroup{}
	wg.Add(len(ids))

	ch := make(chan prometheus.Metric)
	done := make(chan bool)

	for _, id := range ids {
		go m.retrieveMetricsForId(id, ch, wg, labelValues)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	metrics := make([]prometheus.Metric, 0)
	for {
		select {
		case m := <-ch:
			metrics = append(metrics, m)

		case <-done:
			return metrics
		}
	}
}

func (m *StatisticMetricRetriever) retrieveMetricsForId(id string, ch chan<- prometheus.Metric,
	wg *sync.WaitGroup, labelValues map[string][]string) {
	defer wg.Done()

	p := api.StatiscticsPath(m.ressource+"s", id)
	var stats Statistics
	err := m.api.GetAndParse(p, &stats)

	if err != nil {
		log.Errorln(err)
		return
	}

	for _, s := range stats.Statistic {
		if s.Kind == "gauge" {
			ch <- m.convertToMetric(id, &s, labelValues)
		}
	}
}

func (m *StatisticMetricRetriever) convertToMetric(id string, s *Statistic,
	labelValues map[string][]string) prometheus.Metric {
	metricName := strings.Replace(s.Name, ".", "_", -1)

	if s.Unit != "none" {
		metricName += "_" + s.Unit
	}

	n := prometheus.BuildFQName("ovirt", m.ressource, metricName)
	d := prometheus.NewDesc(n, s.Description, m.labelNames, nil)

	return prometheus.MustNewConstMetric(d, prometheus.GaugeValue, s.Values.Value.Datum, labelValues[id]...)
}
