package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/handlers"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

func run() error {
	var MemStorage = storage.NewStore()

	log.Info().Str("Running on", flags.FlagRunAddr).Msg("Server started")
	defer log.Info().Msg("Server stopped")

	mux := chi.NewRouter()
	//mux.Use(middleware.Logger)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.RootHandler(w, r, MemStorage)
	})

	mux.Post("/update/{metricType}/{metricName}/{metricVal}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateHandler(w, r, MemStorage)
	})

	mux.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {
		handlers.ValueHandler(w, r, MemStorage)
	})

	return http.ListenAndServe(flags.FlagRunAddr, mux)
}

func main() {
	flags.ParseFlags()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if err := run(); err != nil {
		panic(err)
	}
}
