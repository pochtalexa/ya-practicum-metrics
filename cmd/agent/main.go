package main

import (
	"errors"
	"fmt"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func initMultiLogger() *os.File {
	fileLogger, err := os.OpenFile(
		"client.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("initMultiLogger")
	}

	writers := io.MultiWriter(os.Stdout, fileLogger)
	log.Logger = log.Output(writers)

	log.Info().Msg("MultiWriter logger initiated")

	return fileLogger
}

func drainChannel(ch <-chan models.Metric, chCashMetricsCapacity int) {
	for i := 0; i < chCashMetricsCapacity; i++ {
		select {
		case <-ch:
		default:
			return
		}
	}

	return
}

func getChanCapacity(runtimeStorage *metrics.RuntimeMetrics, gopsutilStorage *metrics.GopsutilMetrics) (int, error) {
	runtimeMetricsQnty := runtimeStorage.GetMetricsQuantity()
	gopsutilMetricsQnty, err := gopsutilStorage.GetMetricsQuantity()
	if err != nil {
		return -1, fmt.Errorf("getChanCapacity: %w", err)
	}

	return runtimeMetricsQnty + gopsutilMetricsQnty, nil
}

func main() {
	var (
		cashMetrics     metrics.CashMetrics
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

	multiLogger := initMultiLogger()
	defer multiLogger.Close()

	flags.ParseFlags()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	httpClient := http.Client{Transport: tr}

	chCashMetricsCapacity, err := getChanCapacity(runtimeStorage, gopsutilStorage)
	if err != nil {
		log.Fatal().Err(err).Msg("chCashMetricsCapacity")
	}

	chCashMetricsMiddleWare := make(chan models.Metric, chCashMetricsCapacity)
	chCashMetrics := make(chan models.Metric, chCashMetricsCapacity*5)
	chCashMetricsBatch := make(chan metrics.CashMetrics, chCashMetricsCapacity)
	chCashMetricsErrors := make(chan error, flags.FlagWorkers)
	chGopsutilUpdateMetricsResult := make(chan error, 1)

	// создаем пул воркеров
	for i := 0; i < flags.FlagWorkers; i++ {
		workerID := i
		go func() {
			metrics.SendMetricWorker(workerID, chCashMetrics, chCashMetricsErrors, httpClient, flags.FlagRunAddr)
		}()
	}

	// горутина принимаем ошибки от SendMetricWorker и gopsutilStorage.UpdateMetrics
	go func() {
		var err error
		for {
			select {
			case err = <-chCashMetricsErrors:
				log.Info().Err(err).Msg("SendMetricWorker error")
			case err = <-chGopsutilUpdateMetricsResult:
				log.Info().Err(err).Msg("gopsutilStorage.UpdateMetrics error")
			}
		}
	}()

	// горутина читает метрики из chCashMetricsMiddleWare и передает их в chCashMetrics
	// дропает метрики в chCashMetricsMiddleWare если chCashMetrics заполнен
	go func() {
		for v := range chCashMetricsMiddleWare {
			log.Info().Str("len", strconv.Itoa(len(chCashMetricsMiddleWare))).Msg("chCashMetricsMiddleWare")
			log.Info().Str("len", strconv.Itoa(len(chCashMetrics))).Msg("chCashMetrics_len")
			select {
			// try to send value to channel
			case chCashMetrics <- v:
			// if channel is full, drain it
			default:
				log.Info().Err(errors.New("chCashMetrics is full. draining it")).Msg("chCashMetrics")
				drainChannel(chCashMetrics, chCashMetricsCapacity)
			}
		}
	}()

	// горутина: runtimeStorage.UpdateMetrics сбора метрик с заданным интервалом
	wg.Add(1)
	go func() {
		log.Info().Msg("runtimeStorage.UpdateMetrics started")

		for range time.Tick(flags.PollInterval) {
			runtimeStorage.UpdateMetrics()
			log.Info().Msg("runtimeStorage Metrics updated")
		}
		wg.Done()
	}()

	// горутина: gopsutilStorage.UpdateMetrics() сбора метрик с заданным интервалом
	wg.Add(1)
	go func() {
		log.Info().Msg("gopsutilStorage.UpdateMetrics started")

		for range time.Tick(flags.PollInterval) {
			err := gopsutilStorage.UpdateMetrics()
			if err != nil {
				chGopsutilUpdateMetricsResult <- err
			} else {
				log.Info().Msg("gopsutilStorage Metrics updated")
			}
		}
		wg.Done()
	}()

	// горутина: CollectMetrics подготовка кеша для отправки
	// передача кеша в каналы chCashMetricsMiddleWare и chCashMetricsBatch
	wg.Add(1)
	go func() {
		log.Info().Msg("CollectMetrics started")

		for range time.Tick(flags.ReportInterval) {
			cashMetrics, err = metrics.CollectMetrics(runtimeStorage, gopsutilStorage)
			if err != nil {
				log.Fatal().Err(err).Msg("CollectMetrics")
			}
			runtimeStorage.PollCountDrop()
			log.Info().Msg("CollectMetrics done")

			select {
			case chCashMetricsBatch <- cashMetrics:
			default:
				log.Info().
					Err(fmt.Errorf("chCashMetricsBatch is full")).Msg("can not pass metric to channel")
			}

			for _, v := range cashMetrics.CashMetrics {
				// if the channel is full, the default case will be executed
				select {
				case chCashMetricsMiddleWare <- v:
				default:
					log.Info().
						Err(fmt.Errorf("chCashMetricsMiddleWare is full")).
						Str("CashMetric", v.String()).
						Msg("can not pass metric to channel")
				}
			}
		}

		wg.Done()
	}()

	// горутина для отправки метрик батчем
	wg.Add(1)
	go func() {
		for v := range chCashMetricsBatch {
			err = metrics.SendMetricBatch(v, httpClient, flags.FlagRunAddr)
			if err != nil {
				log.Info().Err(err).Msg("SendMetricBatch send error")
			}
		}
		wg.Done()
	}()

	wg.Wait()
}
