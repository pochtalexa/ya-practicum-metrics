package main

import (
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"sync"
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
		log.Fatal().Err(err).Msg("InitMultiLogger")
	}

	writers := io.MultiWriter(os.Stdout, fileLogger)
	log.Logger = log.Output(writers)

	log.Info().Msg("MultiWriter logger initiated")

	return fileLogger
}

func main() {
	var (
		CashMetrics     metrics.CashMetrics
		runtimeStorage  = metrics.NewRuntimeMetrics()
		gopsutilStorage = metrics.NewGopsutilMetrics()
		wg              sync.WaitGroup
		err             error
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

	httpClient := http.Client{Transport: tr}

	chCashMetrics := make(chan models.Metric, 100)
	chCashMetricsResult := make(chan error, flags.FlagWorkers)

	// создаем пул воркеров
	for i := 0; i < flags.FlagWorkers; i++ {
		workerID := i
		go func() {
			metrics.SendMetricWorker(workerID, chCashMetrics, chCashMetricsResult, httpClient, reportRunAddr)
		}()
	}

	// горутина принимаем ошибки от SendMetricWorker
	wg.Add(1)
	go func() {
		for v := range chCashMetricsResult {
			log.Info().Err(v).Msg("SendMetricWorker error")
		}
		wg.Done()
	}()

	// горутина: runtimeStorage.UpdateMetrics сбора метрик с заданным интервалом
	wg.Add(1)
	go func() {
		log.Info().Msg("runtimeStorage.UpdateMetrics started")

		for range time.Tick(time.Second * time.Duration(pollInterval)) {
			runtimeStorage.UpdateMetrics()
			log.Info().Msg("runtimeStorage Metrics updated")
		}
		wg.Done()
	}()

	// горутина: gopsutilStorage.UpdateMetrics() сбора метрик с заданным интервалом
	wg.Add(1)
	go func() {
		log.Info().Msg("gopsutilStorage.UpdateMetrics started")

		for range time.Tick(time.Second * time.Duration(pollInterval)) {
			gopsutilStorage.UpdateMetrics()
			log.Info().Msg("gopsutilStorage Metrics updated")
		}
		wg.Done()
	}()

	// горутина: CollectMetrics подготовка кеша для отправки
	// передача кеша в горутину SendMetricBatch и канал chCashMetrics
	wg.Add(1)
	go func() {
		var wgCollect sync.WaitGroup
		log.Info().Msg("CollectMetrics started")

		for range time.Tick(time.Second * time.Duration(reportInterval)) {
			CashMetrics, err = metrics.CollectMetrics(runtimeStorage, gopsutilStorage)
			if err != nil {
				log.Fatal().Err(err).Msg("CollectMetrics")
			}
			runtimeStorage.PollCountDrop()
			log.Info().Msg("CollectMetrics done")

			for _, v := range CashMetrics.CashMetrics {
				chCashMetrics <- v
			}

			wgCollect.Add(1)
			go func() {
				err = metrics.SendMetricBatch(CashMetrics, httpClient, reportRunAddr)
				if err != nil {
					log.Info().Err(err).Msg("SendMetricBatch send error")
				}
				log.Info().Msg("SendMetricBatch done")
				wgCollect.Done()
			}()

			wgCollect.Wait()
		}

		wg.Done()
	}()

	wg.Wait()
}
