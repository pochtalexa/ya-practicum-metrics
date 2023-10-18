package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

var (
	MemStorage = storage.NewStore()
	DBstorage  = storage.NewDBStore()
	db         *sql.DB
	err        error
)

func catchTermination() {
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	<-shutdownChan
	if flags.StorePoint.File {
		err := storage.StoreMetricsToFile(MemStorage)
		if err != nil {
			panic(err)
		}
	} else if flags.StorePoint.DataBase {
		err := storage.StoreMetricsToDB(DBstorage)
		if err != nil {
			panic(err)
		}
	}

	os.Exit(0)
}

func initStoreTimer() {
	if flags.FlagStoreInterval > 0 {

		for range time.Tick(time.Second * time.Duration(flags.FlagStoreInterval)) {
			if flags.StorePoint.File {
				err := storage.StoreMetricsToFile(MemStorage)
				if err != nil {
					panic(err)
				}
			} else if flags.StorePoint.DataBase {
				err := storage.StoreMetricsToDB(DBstorage)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func restoreMetrics() {
	if flags.FlagRestore && flags.StorePoint.File {
		err := MemStorage.RestoreMetricsFromFile()
		if err != nil {
			log.Info().Err(err).Msg("can not read metrics from file")
			return
		}
		log.Info().Msg("metrics restored from file")
	} else if flags.FlagRestore && flags.StorePoint.DataBase {
		err := DBstorage.RestoreMetricsFromDB()
		if err != nil {
			log.Info().Err(err).Msg("can not read metrics from DB")
			return
		}
	}
}

func run() error {

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	// return all metrics on WEB page
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if flags.StorePoint.DataBase {
			handlers.RootHandler(w, r, DBstorage)
		} else {
			handlers.RootHandler(w, r, MemStorage)
		}
	})

	// ping DB
	mux.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		handlers.PingHandler(w, r, db)
	})

	// get metrics in array
	mux.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdatesHandler(w, r, DBstorage)
	})

	mux.Post("/update/{metricType}/{metricName}/{metricVal}", func(w http.ResponseWriter, r *http.Request) {
		if flags.StorePoint.DataBase {
			handlers.UpdateHandlerLong(w, r, DBstorage)
		} else {
			handlers.UpdateHandlerLong(w, r, MemStorage)
		}
	})

	mux.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
		if flags.StorePoint.DataBase {
			handlers.UpdateHandler(w, r, DBstorage)
		} else {
			handlers.UpdateHandler(w, r, MemStorage)
		}
	})

	mux.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {
		if flags.StorePoint.DataBase {
			handlers.ValueHandlerLong(w, r, DBstorage)
		} else {
			handlers.ValueHandlerLong(w, r, MemStorage)
		}
	})

	mux.Post("/value/", func(w http.ResponseWriter, r *http.Request) {
		if flags.StorePoint.DataBase {
			handlers.ValueHandler(w, r, DBstorage)
		} else {
			handlers.ValueHandler(w, r, MemStorage)
		}
	})

	log.Info().Str("Running on", flags.FlagRunAddr).Msg("Server started")
	defer log.Info().Msg("Server stopped")

	return http.ListenAndServe(flags.FlagRunAddr, mux)
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	flags.ParseFlags()

	if flags.StorePoint.DataBase {
		db, err = storage.InitConnDB()
		if err != nil {
			log.Err(err)
			panic(err)
		}
		defer db.Close()

		err := storage.InitialazeDB(db)
		if err != nil {
			log.Err(err)
			panic(err)
		}

		DBstorage.DBconn = db
	}

	go catchTermination()
	restoreMetrics()
	go initStoreTimer()

	if err := run(); err != nil {
		panic(err)
	}
}
