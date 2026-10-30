package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"httpServer/internal/http1"
	"httpServer/internal/mux"

	"github.com/VictoriaMetrics/metrics"
)

func main() {
	m := mux.NewSimpleMux()

	m.Use(mux.Logger)
	m.Use(mux.Recoverer)
	m.Use(mux.MetricsAndTraceMiddleware)

	m.HandleFunc("GET /app-stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		metrics.WritePrometheus(w, true)
	})

	m.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "The test is successfull")
	})

	m.HandleFunc("GET /delay", func(w http.ResponseWriter, r *http.Request) {
		sleepMs := 50 + rand.Intn(350)
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		fmt.Fprintf(w, "Delayed response after %dms\n", sleepMs)
	})

	m.HandleFunc("GET /flaky", func(w http.ResponseWriter, r *http.Request) {
		if rand.Float32() < 0.3 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "500 Internal Server Error (Chaos triggered)")
			return
		}
		fmt.Fprintln(w, "200 OK - Request succeeded")
	})

	m.HandleFunc("GET /payload", func(w http.ResponseWriter, r *http.Request) {
		repeat := 10 + rand.Intn(90) // 10 bis 100 KB
		data := strings.Repeat("A", repeat*1024)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "Generated payload size: %d KB\n%s", repeat, data)
	})

	m.HandleFunc("GET /{catchall}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Mux wasnt able to find the path: %q\n", r.URL.Path)
	})

	if err := http1.ListenAndServe(":8080", m); err != nil {
		log.Fatal(err)
	}
}
