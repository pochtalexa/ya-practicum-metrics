package main

import (
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var (
	pollInterval   int
	reportInterval int
	reportRunAddr  string
)

func main() {
	var (
		CashMetrics    metrics.CashMetrics
		metricsStorage = metrics.New()
		err            error
	)

	flags.ParseFlags()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	pollInterval = flags.FlagPollInterval
	reportInterval = flags.FlagReportInterval
	reportRunAddr = flags.FlagRunAddr

	pollIntervalCounter := 0
	reportIntervalCounter := 0

	httpClient := http.Client{}

	for {
		time.Sleep(time.Duration(1) * time.Second)

		pollIntervalCounter += 1
		reportIntervalCounter += 1

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

			err = metrics.SendMetric(CashMetrics, httpClient, reportRunAddr)
			if err != nil {
				panic(err)
			}

			reportIntervalCounter = 0
			metricsStorage.PollCountDrop()

			log.Info().Msg("Metrics sent")
		}
	}
}
