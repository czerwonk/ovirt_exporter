// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/czerwonk/ovirt_api/api"
	"github.com/czerwonk/ovirt_exporter/pkg/collector.go"
	"github.com/czerwonk/ovirt_exporter/pkg/host"
	"github.com/czerwonk/ovirt_exporter/pkg/storagedomain"
	"github.com/czerwonk/ovirt_exporter/pkg/vm"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const version string = "0.10.0"

var (
	showVersion              = flag.Bool("version", false, "Print version information.")
	listenAddress            = flag.String("web.listen-address", ":9325", "Address on which to expose metrics and web interface.")
	metricsPath              = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	apiURL                   = flag.String("api.url", "https://localhost/ovirt-engine/api/", "API REST Endpoint")
	apiUser                  = flag.String("api.username", "user@internal", "API username")
	apiPass                  = flag.String("api.password", "", "API password")
	apiPassFile              = flag.String("api.password-file", "", "File containing the API password")
	apiInsecureCert          = flag.Bool("api.insecure-cert", false, "Skip verification for untrusted SSL/TLS certificates")
	withSnapshots            = flag.Bool("with-snapshots", true, "Collect snapshot metrics (can be time consuming in some cases)")
	withNetwork              = flag.Bool("with-network", true, "Collect network metrics (can be time consuming in some cases)")
	withDisks                = flag.Bool("with-disks", true, "Collect disk metrics (can be time consuming in some cases)")
	debug                    = flag.Bool("debug", false, "Show verbose output (e.g. body of each response received from API)")
	tlsEnabled               = flag.Bool("tls.enabled", false, "Enables TLS")
	tlsCertChainPath         = flag.String("tls.cert-file", "", "Path to TLS cert file")
	tlsKeyPath               = flag.String("tls.key-file", "", "Path to TLS key file")
	tracingEnabled           = flag.Bool("tracing.enabled", false, "Enables tracing using OpenTelemetry")
	tracingProvider          = flag.String("tracing.provider", "", "Sets the tracing provider (stdout or collector)")
	tracingCollectorEndpoint = flag.String("tracing.collector.grpc-endpoint", "", "Sets the tracing provider (stdout or collector)")

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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdownTracing, err := initTracing(ctx)
	if err != nil {
		log.Fatalf("could not initialize tracing: %v", err)
	}
	defer shutdownTracing()

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
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectorDuration)

	http.HandleFunc(*metricsPath, func(w http.ResponseWriter, r *http.Request) {
		handleMetricsRequest(w, r, client, reg)
	})

	log.Infof("Listening for %s on %s (TLS: %v)", *metricsPath, *listenAddress, *tlsEnabled)
	if *tlsEnabled {
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertChainPath, *tlsKeyPath, nil))
		return
	}

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func connectAPI() (*api.Client, error) {
	opts := []api.ClientOption{api.WithLogger(log.StandardLogger())}

	if *debug {
		opts = append(opts, api.WithDebug())
	}

	if *apiInsecureCert {
		opts = append(opts, api.WithInsecure())
	}

	pass, err := apiPassword()
	if err != nil {
		return nil, errors.Wrap(err, "error while reading password file")
	}

	client, err := api.NewClient(*apiURL, *apiUser, pass, opts...)
	if err != nil {
		return nil, err
	}

	return client, err
}

func apiPassword() (string, error) {
	if *apiPassFile == "" {
		return *apiPass, nil
	}

	b, err := os.ReadFile(*apiPassFile)
	if err != nil {
		return "", err
	}

	return strings.Trim(string(b), "\n"), nil
}

func handleMetricsRequest(w http.ResponseWriter, r *http.Request, client *api.Client, appReg *prometheus.Registry) {
	ctx, span := tracer.Start(context.Background(), "HandleMetricsRequest")
	defer span.End()

	reg := prometheus.NewRegistry()

	cc := collector.NewContext(tracer, client)
	reg.MustRegister(vm.NewCollector(ctx, cc.Clone(), *withSnapshots, *withNetwork, *withDisks, collectorDuration.WithLabelValues("vm")))
	reg.MustRegister(host.NewCollector(ctx, cc.Clone(), *withNetwork, collectorDuration.WithLabelValues("host")))
	reg.MustRegister(storagedomain.NewCollector(ctx, cc.Clone(), collectorDuration.WithLabelValues("storage")))

	multiRegs := prometheus.Gatherers{
		reg,
		appReg,
	}

	l := log.New()
	l.Level = log.ErrorLevel

	promhttp.HandlerFor(multiRegs, promhttp.HandlerOpts{
		ErrorLog:      l,
		ErrorHandling: promhttp.ContinueOnError,
		Registry:      appReg}).ServeHTTP(w, r)
}
