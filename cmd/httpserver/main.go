package main

import (
	"fmt"
	"httpServer/internal/http1"
	"httpServer/internal/mux"
	"net/http"
)

const addr = ":8080"

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})

	newMux()

	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func newMux() {

	mux := mux.NewSimpleMux()

	mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "The test is successfull")
	})

	http.ListenAndServe(":8088", mux)
}
