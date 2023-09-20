package main

import (
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/send"
	"runtime"
	"time"
)

var (
	pollInterval   int
	reportInterval int
	reportRunAddr  string
)

func main() {
	flags.ParseFlags()

	pollInterval = flags.FlagPollInterval
	reportInterval = flags.FlagReportInterval
	reportRunAddr = flags.FlagRunAddr

	pollIntervalCounter := 0
	reportIntervalCounter := 0

	metricsStorage := metrics.New()

	for {
		time.Sleep(time.Duration(1) * time.Second)

		pollIntervalCounter += 1
		reportIntervalCounter += 1

		if pollIntervalCounter == pollInterval {
			runtime.ReadMemStats(&metricsStorage.Data)
			metricsStorage.RandomValueUpdate()
			metricsStorage.PollCountInc()
			pollIntervalCounter = 0
		}

		if reportIntervalCounter == reportInterval {
			err := send.Metrics(metricsStorage, reportRunAddr)
			if err != nil {
				panic(err)
			}
			reportIntervalCounter = 0
			metricsStorage.PollCountDrop()
		}
	}
}
