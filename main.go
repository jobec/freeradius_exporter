package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bvantagelimited/freeradius_exporter/client"
	"github.com/bvantagelimited/freeradius_exporter/collector"
)

func main() {
	listenAddr := flag.String("web.listen-address", ":9812", "Address to listen on for web interface and telemetry.")
	metricsPath := flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	radiusTimeout := flag.Int("radius.timeout", 5000, "Timeout, in milliseconds.")
	radiusAddr := flag.String("radius.address", getEnv("RADIUS_ADDR", "127.0.0.1:18121"), "Address of FreeRADIUS status server.")
	radiusSecret := flag.String("radius.secret", getEnv("RADIUS_SECRET", "testing123"), "FreeRADIUS client secret.")

	flag.Parse()

	registry := prometheus.NewRegistry()

	radiusClient, err := client.NewFreeRADIUSClient(*radiusAddr, *radiusSecret, *radiusTimeout)
	if err != nil {
		log.Fatal(err)
	}

	registry.MustRegister(collector.NewFreeRADIUSCollector(radiusClient))

	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>FreeRADIUS Exporter</title></head>
			<body>
			<h1>FreeRADIUS Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})

	srv := &http.Server{}
	listener, err := net.Listen("tcp4", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Providing metrics at %s%s", *listenAddr, *metricsPath)
	log.Fatal(srv.Serve(listener))
}

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}
