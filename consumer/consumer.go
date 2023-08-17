package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := os.Setenv("TZ", ""); err != nil {
		return
	}

	latencies := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "event_delivery_latency_seconds",
		Help:    "The latency of each event delivery",
		Buckets: prometheus.LinearBuckets(1.0, 1.5, 10),
	})

	reg := prometheus.NewRegistry()
	reg.MustRegister(latencies)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	mux.HandleFunc("/none", func(w http.ResponseWriter, req *http.Request) {
		// record metric.
		if timeHeader, found := req.Header["X-Benchmark-Timestamp"]; found {
			if len(timeHeader) != 1 {
				// end.
				w.Write([]byte("End."))
				return
			}

			st, err := strconv.ParseInt(timeHeader[0], 10, 64)
			if err != nil {
				// end.
				w.Write([]byte("End."))
				return
			}

			ft := time.Now()
			latency := ft.Sub(time.Unix(st, 0))
			latencies.Observe(latency.Seconds())

			w.Write([]byte("Great."))
		}
	})

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(srv.ListenAndServe())
}
