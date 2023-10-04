package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/handlers"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var MemStorage = storage.NewStore()

func catchTermination() {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	<-shutdownChan
	_ = MemStorage.StoreMetricsToFile()

	os.Exit(0)
}

func initStoreTimer() {
	if flags.FlagStoreInterval > 0 {

		for range time.Tick(time.Second * time.Duration(flags.FlagStoreInterval)) {
			err := MemStorage.StoreMetricsToFile()
			if err != nil {
				panic(err)
			}
		}
	}
}

func restoreMetrics() {
	if flags.FlagRestore {
		err := MemStorage.RestoreMetricsFromFile()
		if err != nil {
			log.Info().Err(err).Msg("can not read metrics from file")
			return
		}
		log.Info().Msg("metrics restored from file")
	}
}

func run() error {

	mux := chi.NewRouter()
	//mux.Use(middleware.Logger)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.RootHandler(w, r, MemStorage)
	})

	mux.Post("/update/{metricType}/{metricName}/{metricVal}", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateHandlerLong(w, r, MemStorage)
	})
	mux.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateHandler(w, r, MemStorage)
	})

	mux.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {
		handlers.ValueHandlerLong(w, r, MemStorage)
	})
	mux.Post("/value/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ValueHandler(w, r, MemStorage)
	})

	log.Info().Str("Running on", flags.FlagRunAddr).Msg("Server started")
	defer log.Info().Msg("Server stopped")

	return http.ListenAndServe(flags.FlagRunAddr, mux)
}

func main() {
	flags.ParseFlags()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	go catchTermination()

	restoreMetrics()
	go initStoreTimer()
	if err := run(); err != nil {
		panic(err)
	}
}
