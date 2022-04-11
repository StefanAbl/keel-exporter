package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stefanabl/keel-exporter/collector"
)

var (
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

	keelUrl  = os.Getenv("KEEL_URL")
	keelUser = os.Getenv("KEEL_USER")
	keelPass = os.Getenv("KEEL_PASS")
)

func main() {
	var r = prometheus.NewRegistry()
	var collector = collector.NewKeelCollector(keelUrl, keelUser, keelPass)
	r.MustRegister(collector)
	flag.Parse()
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
