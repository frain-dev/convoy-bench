package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	latencies := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "event_delivery_seconds",
		Help:    "The latency in seconds for each event delivery",
		Buckets: prometheus.LinearBuckets(1, 1, 100),
	})

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration distribution",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(latencies)
	reg.MustRegister(httpRequestsTotal)
	reg.MustRegister(httpRequestDuration)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	mux.HandleFunc("/none", func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

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

			elapsed := time.Since(start).Seconds()

			// Increment request count and record request duration.
			httpRequestsTotal.WithLabelValues(req.Method).Inc()
			httpRequestDuration.WithLabelValues(req.Method).Observe(elapsed)

			w.Write([]byte("Great."))
		}
	})

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Fatal(srv.ListenAndServe())
}
