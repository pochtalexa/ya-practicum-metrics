package main

import (
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	pollInterval   int
	reportInterval int
	reportRunAddr  string
)

func InitMultiLogger() *os.File {
	fileLogger, err := os.OpenFile(
		"client.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}

	writers := io.MultiWriter(os.Stdout, fileLogger)
	log.Logger = log.Output(writers)

	log.Info().Msg("MultiWriter logger initiated")

	return fileLogger
}

func main() {
	var (
		CashMetrics    metrics.CashMetrics
		metricsStorage = metrics.New()
		err            error
	)

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	multiLogger := InitMultiLogger()
	defer multiLogger.Close()

	flags.ParseFlags()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	pollInterval = flags.FlagPollInterval
	reportInterval = flags.FlagReportInterval
	reportRunAddr = flags.FlagRunAddr

	pollIntervalCounter := 0
	reportIntervalCounter := 0

	httpClient := http.Client{Transport: tr}

	for {
		time.Sleep(1 * time.Second)

		pollIntervalCounter++
		reportIntervalCounter++

		if pollIntervalCounter == pollInterval {
			metricsStorage.UpdateMetrics()
			pollIntervalCounter = 0

			log.Info().Msg("Metrics updated")
		}

		if reportIntervalCounter == reportInterval {
			CashMetrics, err = metrics.CollectMetrics(metricsStorage)
			if err != nil {
				panic(err)
			}
			//err = metrics.SendMetric(CashMetrics, httpClient, reportRunAddr)
			err = metrics.SendMetricBatch(CashMetrics, httpClient, reportRunAddr)
			if err != nil {
				panic(err)
			}
			reportIntervalCounter = 0
			metricsStorage.PollCountDrop()
			log.Info().Msg("Metrics sent")

		}
	}
}
