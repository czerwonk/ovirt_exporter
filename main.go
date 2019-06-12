package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/czerwonk/ovirt_api/api"
	"github.com/czerwonk/ovirt_exporter/host"
	"github.com/czerwonk/ovirt_exporter/storagedomain"
	"github.com/czerwonk/ovirt_exporter/vm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const version string = "0.8.6"

var (
	showVersion     = flag.Bool("version", false, "Print version information.")
	listenAddress   = flag.String("web.listen-address", ":9325", "Address on which to expose metrics and web interface.")
	metricsPath     = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	apiURL          = flag.String("api.url", "https://localhost/ovirt-engine/api/", "API REST Endpoint")
	apiUser         = flag.String("api.username", "user@internal", "API username")
	apiPass         = flag.String("api.password", "", "API password")
	apiInsecureCert = flag.Bool("api.insecure-cert", false, "Skip verification for untrusted SSL/TLS certificates")
	withSnapshots   = flag.Bool("with-snapshots", true, "Collect snapshot metrics (can be time consuming in some cases)")
	withNetwork     = flag.Bool("with-network", true, "Collect network metrics (can be time consuming in some cases)")
	debug           = flag.Bool("debug", false, "Show verbose output (e.g. body of each response received from API)")

	collectorDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ovirt_collectors_duration",
			Help:    "Histogram of latencies for metric collectors.",
			Buckets: []float64{.1, .2, .4, 1, 3, 8, 20, 60},
		},
		[]string{"collector"},
	)
)

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: ovirt_exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	startServer()
}

func printVersion() {
	fmt.Println("ovirt_exporter")
	fmt.Printf("Version: %s\n", version)
	fmt.Println("Author(s): Daniel Czerwonk")
	fmt.Println("Metric exporter for oVirt engine")
}

func startServer() {
	log.Infof("Starting oVirt exporter (Version: %s)", version)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>oVirt Exporter (Version ` + version + `)</title></head>
			<body>
			<h1>oVirt Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/czerwonk/ovirt_exporter">github.com/czerwonk/ovirt_exporter</a></p>
			</body>
			</html>`))
	})

	client, err := connectAPI()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(collectorDuration)

	http.HandleFunc(*metricsPath, func(w http.ResponseWriter, r *http.Request) {
		handleMetricsRequest(w, r, client, reg)
	})

	log.Infof("Listening for %s on %s", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func connectAPI() (*api.Client, error) {
	opts := []api.ClientOption{api.WithLogger(&PromLogger{})}

	if *debug {
		opts = append(opts, api.WithDebug())
	}

	if *apiInsecureCert {
		opts = append(opts, api.WithInsecure())
	}

	client, err := api.NewClient(*apiURL, *apiUser, *apiPass, opts...)
	if err != nil {
		return nil, err
	}

	return client, err
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request, client *api.Client, appReg *prometheus.Registry) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(vm.NewCollector(client, *withSnapshots, *withNetwork, collectorDuration.WithLabelValues("vm")))
	reg.MustRegister(host.NewCollector(client, *withNetwork, collectorDuration.WithLabelValues("host")))
	reg.MustRegister(storagedomain.NewCollector(client, collectorDuration.WithLabelValues("storage")))

	multiRegs := prometheus.Gatherers{
		reg,
		appReg,
	}

	promhttp.HandlerFor(multiRegs, promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError,
		Registry:      appReg}).ServeHTTP(w, r)
}
