package mux

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type traceKey struct{}

func MetricsAndTraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		traceID := r.Header.Get("traceparent")
		if traceID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			traceID = fmt.Sprintf("00-%s-0000000000000001-01", hex.EncodeToString(b))
		}

		ctx := context.WithValue(r.Context(), traceKey{}, traceID)
		w.Header().Set("traceparent", traceID)

		next.ServeHTTP(w, r.WithContext(ctx))

		duration := time.Since(start).Seconds()

		statusCode := http.StatusOK
		bytesWritten := 0
		if customResp, ok := w.(interface {
			StatusCode() int
			BytesWritten() int
		}); ok {
			statusCode = customResp.StatusCode()
			bytesWritten = customResp.BytesWritten()
		}

		statusStr := strconv.Itoa(statusCode)
		path := r.URL.Path

		metrics.GetOrCreateCounter(fmt.Sprintf(`http_requests_total{path="%s",method="%s",status="%s"}`, path, r.Method, statusStr)).Inc()

		metrics.GetOrCreateHistogram(fmt.Sprintf(`http_request_duration_seconds{path="%s",method="%s"}`, path, r.Method)).Update(duration)

		metrics.GetOrCreateCounter(fmt.Sprintf(`http_response_bytes_total{path="%s"}`, path)).Add(bytesWritten)
	})
}
