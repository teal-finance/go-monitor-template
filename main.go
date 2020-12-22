package main

import (
	"log"
	"net/http"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func monitorRED(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		record := &statusRecorder{ResponseWriter: w, Status: "success"}

		next.ServeHTTP(record, r)

		labels := []metrics.Label{{Name: "method", Value: r.Method}, {Name: "route", Value: r.RequestURI}, {Name: "status", Value: record.Status}}

		duration := time.Since(start)
		metrics.AddSampleWithLabels([]string{"request_duration"}, float32(duration.Milliseconds()), labels)

		log.Printf("%v %v %v\n", r.Method, r.RequestURI, duration)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	Status string
}

func (r *statusRecorder) WriteHeader(status int) {
	if status != http.StatusOK {
		r.Status = "error"
	}
	r.ResponseWriter.WriteHeader(status)
}

func main() {
	mon, _ := prometheus.NewPrometheusSink()
	metrics.NewGlobal(metrics.DefaultConfig("myapp"), mon)

	// rest service with prometheus endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write([]byte(`hello`))
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: monitorRED(mux),
	}
	log.Println("Starting server at", server.Addr)
	log.Fatal(server.ListenAndServe())
}
