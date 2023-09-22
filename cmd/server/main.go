package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/handlers"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
	"net/http"
)

var MemStorage = storage.NewStore()

func run() error {
	fmt.Println("Running server on", flags.FlagRunAddr)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.RootHandler(w, r, MemStorage)
	})

	mux.Post("/update/{metricType}/{metricName}/{metricVal}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateHandler(w, r, *MemStorage)
	})

	mux.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {
		handlers.ValueHandler(w, r, *MemStorage)
	})

	return http.ListenAndServe(flags.FlagRunAddr, mux)
}

func main() {
	flags.ParseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}
