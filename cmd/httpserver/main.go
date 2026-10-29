package main

import (
	"fmt"
	"log"
	"net/http"

	"httpServer/internal/http1"
	"httpServer/internal/mux"

	"github.com/VictoriaMetrics/metrics"
)

func main() {
	m := mux.NewSimpleMux()

	m.Use(mux.Logger)
	m.Use(mux.Recoverer)

	m.HandleFunc("GET /telemetry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		metrics.WritePrometheus(w, true)
	})

	m.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		metrics.GetOrCreateCounter(fmt.Sprintf(`http_requests_total{path="%s",method="%s"}`, "/test", r.Method)).Inc()
		_, _ = fmt.Fprintln(w, "The test is successfull")
	})

	if err := http1.ListenAndServe(":8080", m); err != nil {
		log.Fatal(err)
	}
}
