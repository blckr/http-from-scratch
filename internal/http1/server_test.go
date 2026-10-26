package http1

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestHandleConnection(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})
	mux.HandleFunc("/custom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "Golang")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "custom accepted")
	})

	tests := []struct {
		name            string
		request         string
		expectedStatus  string
		expectedHeaders []string
		expectedBody    string
	}{
		{
			name:           "Standard 200 OK",
			request:        "GET /hello HTTP/1.1\r\nHost: localhost:8080\r\nConnection: close\r\n\r\n",
			expectedStatus: "HTTP/1.1 200 OK",
			expectedBody:   "hello",
		},
		{
			name:           "404 Route Not Found",
			request:        "GET /unknown-path HTTP/1.1\r\nHost: localhost:8080\r\nConnection: close\r\n\r\n",
			expectedStatus: "HTTP/1.1 404 Not Found",
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "202 Accepted with Header",
			request:        "GET /custom HTTP/1.1\r\nHost: localhost:8080\r\nConnection: close\r\n\r\n",
			expectedStatus: "HTTP/1.1 202 Accepted",
			expectedHeaders: []string{
				"X-Custom-Header: Golang",
			},
			expectedBody: "custom accepted",
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {

			// Create a pipe to perform the conservation
			client, server := net.Pipe()

			go func() {
				_ = HandleConnection(server, mux)
			}()

			if _, err := client.Write([]byte(tc.request)); err != nil {
				t.Fatalf("failed to send request: %v", err)
			}

			data, err := io.ReadAll(client)
			if err != nil {
				t.Fatalf("failed to read response: %v", err)
			}
			_ = client.Close()

			response := string(data)

			if !strings.Contains(response, tc.expectedStatus) {
				t.Errorf("expected status %q, got:\n%s", tc.expectedStatus, response)
			}

			for _, h := range tc.expectedHeaders {
				if !strings.Contains(response, h) {
					t.Errorf("expected header %q, got:\n%s", h, response)
				}
			}

			if !strings.HasSuffix(response, tc.expectedBody) {
				t.Errorf("expected body to end with %q, got:\n%s", tc.expectedBody, response)
			}
		})
	}
}
