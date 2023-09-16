package main

import (
	"net/http"

	"github.com/pochtalexa/ya-practicum-metrics/internal/server/handlers"
)

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/update/", handlers.UpdateHandler)
	mux.HandleFunc("/value/", handlers.ValueHandler)
	return http.ListenAndServe(":8080", mux)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
