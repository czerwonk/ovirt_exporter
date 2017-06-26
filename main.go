package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/czerwonk/ovirt_exporter/host"
	"github.com/czerwonk/ovirt_exporter/storagedomain"
	"github.com/czerwonk/ovirt_exporter/vm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

const version string = "0.2.0"

var (
	showVersion     = flag.Bool("version", false, "Print version information.")
	listenAddress   = flag.String("web.listen-address", ":9325", "Address on which to expose metrics and web interface.")
	metricsPath     = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	apiUrl          = flag.String("api.url", "https://localhost/ovirt-engine/api/", "API REST Endpoint")
	apiUser         = flag.String("api.username", "user@internal", "API username")
	apiPass         = flag.String("api.password", "", "API password")
	apiInsecureCert = flag.Bool("api.insecure-cert", false, "Skip verification for untrusted SSL/TLS certificates")
	client          *api.ApiClient
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
	log.Infof("Starting oVirt exporter (Version: %s)\n", version)
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
	http.HandleFunc(*metricsPath, handleMetricsRequest)

	client = api.NewClient(*apiUrl, *apiUser, *apiPass, *apiInsecureCert)

	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(vm.NewCollector(client))
	reg.MustRegister(host.NewCollector(client))
	reg.MustRegister(storagedomain.NewCollector(client))

	promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      log.NewErrorLogger(),
		ErrorHandling: promhttp.ContinueOnError}).ServeHTTP(w, r)
}
