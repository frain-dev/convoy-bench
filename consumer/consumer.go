package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gonum.org/v1/gonum/stat"
)

func main() {
	mutex := sync.Mutex{}
	// Atomic so the delivery handlers never take a lock: one per delivery would
	// have the consumer contending with itself at the rates it is measuring.
	// Everything off that hot path (the sampler, /clear, the reporting handlers)
	// still goes through the mutex, which is what keeps reqs and rps consistent
	// with each other.
	var reqs atomic.Int64
	rps := make([]int, 0)
	type Response struct {
		Data []int `json:"data"`
	}

	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			// Swap and append under one lock. Splitting them lets /clear wipe rps
			// in between, after which this appends a sample from the window that
			// was just cleared and the next run starts dirty.
			mutex.Lock()
			if n := reqs.Swap(0); n != 0 {
				rps = append(rps, int(n))
			}
			mutex.Unlock()
		}
	}()

	latencies := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "event_delivery_seconds",
		Help: "The latency in seconds for each event delivery",
		// 1ms doubling to roughly 8.7 minutes. Buckets must start well below a
		// second, or every sub-second delivery collapses into one bucket and the
		// histogram cannot distinguish a fast cluster from a slow one.
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 20),
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
		// make+copy rather than append to a nil slice, so an empty window still
		// marshals as [] and not null.
		mutex.Lock()
		snapshot := make([]int, len(rps))
		copy(snapshot, rps)
		mutex.Unlock()

		res := Response{
			Data: snapshot,
		}
		b, _ := json.Marshal(res)

		_, _ = w.Write(b)
	})

	mux.HandleFunc("/clear", func(w http.ResponseWriter, req *http.Request) {
		mutex.Lock()
		rps = []int{}
		reqs.Store(0)
		mutex.Unlock()

		res := Response{
			Data: []int{},
		}
		b, _ := json.Marshal(res)

		_, _ = w.Write(b)
	})

	mux.HandleFunc("/px", func(w http.ResponseWriter, req *http.Request) {
		mutex.Lock()
		data := make([]float64, len(rps))
		for i, v := range rps {
			data[i] = float64(v)
		}
		mutex.Unlock()
		sort.Float64s(data)
		mean := stat.Mean(data, nil)
		p95 := percentile(data, 95)
		p99 := percentile(data, 99)
		p1 := percentile(data, 1)
		p5 := percentile(data, 5)

		// len(data), not len(rps): the count has to describe the same snapshot the
		// percentiles were computed from, and rps can grow between the two reads.
		display := fmt.Sprintf("Count: %d\nMean: %.2f\np(1): %.2f\np(5): %.2f\np(95): %.2f\np(99): %.2f\n", len(data), mean, p1, p5, p95, p99)
		_, _ = w.Write([]byte(display))
	})

	mux.HandleFunc("/none", func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()

		// record metric.
		// Reject rather than accept an unusable timestamp: a run that measured
		// nothing must not be mistaken for a run that measured zero latency.
		if timeHeader, found := req.Header["X-Benchmark-Timestamp-Ms"]; found {
			if len(timeHeader) != 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Expected exactly one X-Benchmark-Timestamp-Ms header"))
				return
			}

			st, err := strconv.ParseInt(timeHeader[0], 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("X-Benchmark-Timestamp-Ms is not an integer"))
				return
			}

			ft := time.Now()
			latency := ft.Sub(time.UnixMilli(st))
			latencies.Observe(latency.Seconds())

			elapsed := time.Since(start).Seconds()

			// Increment request count and record request duration.
			httpRequestDuration.Observe(elapsed)
			httpRequestsTotal.WithLabelValues(req.Method).Inc()

			reqs.Add(1)

			_, _ = w.Write([]byte("Great."))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Required header X-Benchmark-Timestamp-Ms missing"))
	})

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Println("running on port 8080")
	log.Fatal(srv.ListenAndServe())
}

// percentile calculates the value at a given percentile using the nearest rank method.
func percentile(data []float64, perc float64) float64 {
	if len(data) == 0 {
		return math.NaN()
	}
	k := float64(len(data)-1) * (perc / 100.0)
	f := math.Floor(k)
	c := math.Ceil(k)
	if f == c {
		return data[int(k)]
	}
	d0 := data[int(f)] * (c - k)
	d1 := data[int(c)] * (k - f)
	return d0 + d1
}
