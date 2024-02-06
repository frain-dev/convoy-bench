package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var reqs, rps = 0, make([]int, 0)
	type Response struct {
		Data []int `json:"data"`
	}

	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			if reqs != 0 {
				rps = append(rps, reqs)
			}
			reqs = 0
		}
	}()

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

	httpRequestDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration distribution",
		Buckets: prometheus.DefBuckets,
	})

	reg := prometheus.NewRegistry()
	reg.MustRegister(latencies)
	reg.MustRegister(httpRequestsTotal)
	reg.MustRegister(httpRequestDuration)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	mux.HandleFunc("/rps", func(w http.ResponseWriter, req *http.Request) {
		res := Response{
			Data: rps,
		}
		b, _ := json.Marshal(res)

		_, _ = w.Write(b)
	})

	mux.HandleFunc("/clear", func(w http.ResponseWriter, req *http.Request) {
		rps = []int{}
		reqs = 0

		res := Response{
			Data: rps,
		}
		b, _ := json.Marshal(res)

		_, _ = w.Write(b)
	})

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
			httpRequestDuration.Observe(elapsed)
			httpRequestsTotal.WithLabelValues(req.Method).Inc()

			reqs++

			w.Write([]byte("Great."))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Required header X-Benchmark-Timestamp missing"))
	})

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Println("running on port 8080")
	log.Fatal(srv.ListenAndServe())
}
