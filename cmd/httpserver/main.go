package main

import (
	"fmt"
	"log"
	"net/http"

	"httpServer/internal/http1"
	"httpServer/internal/mux"
)

func main() {
	m := mux.NewSimpleMux()

	m.Use(mux.Logger)
	m.Use(mux.Recoverer)

	m.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "The test is successfull")
	})

	if err := http1.ListenAndServe(":8088", m); err != nil {
		log.Fatal(err)
	}
}
