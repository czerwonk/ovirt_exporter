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
	ressource string
	api       *api.ApiClient
}

func NewStatisticMetricRetriever(ressource string, api *api.ApiClient) *StatisticMetricRetriever {
	return &StatisticMetricRetriever{ressource: ressource, api: api}
}

func (m *StatisticMetricRetriever) RetrieveMetrics(ressources map[string]string) []prometheus.Metric {
	wg := &sync.WaitGroup{}
	wg.Add(len(ressources))

	ch := make(chan prometheus.Metric)
	done := make(chan bool)

	for id, name := range ressources {
		go m.retrieveMetricsForId(id, name, ch, wg)
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

func (m *StatisticMetricRetriever) retrieveMetricsForId(id string, name string, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
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
			ch <- m.convertToMetric(id, name, &s)
		}
	}
}

func (m *StatisticMetricRetriever) convertToMetric(id string, name string, s *Statistic) prometheus.Metric {
	metricName := strings.Replace(s.Name, ".", "_", -1)

	if s.Unit != "none" {
		metricName += "_" + s.Unit
	}

	n := prometheus.BuildFQName("ovirt", m.ressource, metricName)

	labelNames := []string{"name"}
	d := prometheus.NewDesc(n, s.Description, labelNames, nil)

	r, err := prometheus.NewConstMetric(d, prometheus.GaugeValue, s.Values.Value.Datum, name)

	if err != nil {
		log.Errorln(err)
	}

	return r
}
