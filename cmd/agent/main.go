package main

import (
	patronhttp "github.com/beatlabs/patron/client/http"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
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

	pollInterval = flags.FlagPollInterval
	reportInterval = flags.FlagReportInterval
	reportRunAddr = flags.FlagRunAddr

	pollIntervalCounter := 0
	reportIntervalCounter := 0

	httpClient, err := patronhttp.New()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Duration(1) * time.Second)

		pollIntervalCounter += 1
		reportIntervalCounter += 1

		if pollIntervalCounter == pollInterval {
			metricsStorage.UpdateMetrics()
			pollIntervalCounter = 0
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
		}
	}
}
