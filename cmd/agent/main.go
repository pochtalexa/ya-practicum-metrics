package main

import (
	"fmt"
	patronhttp "github.com/beatlabs/patron/client/http"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"golang.org/x/exp/slices"
)

var (
	rtm            runtime.MemStats
	pollInterval   int
	reportInterval int
	reportRunAddr  string
)

const (
// pollInterval   = 2
// reportInterval = 10
// reportHost     = "127.0.0.1"
// reportPort     = "8080"
)

func sendMetrics(metrics *metrics.RuntimeMetrics) error {
	var urlGauge string

	//httpClient := &http.Client{}
	httpClient, err := patronhttp.New()
	if err != nil {
		return err
	}

	fields := reflect.TypeOf(rtm)
	values := reflect.ValueOf(rtm)
	num := fields.NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)

		if slices.Contains(metrics.MetricsName, field.Name) {
			//fmt.Println("Type:", field.Type, ",", field.Name, "=", value)

			if value.Kind() == reflect.Float64 {
				urlGauge = fmt.Sprintf("http://%s/update/gauge/%s/%.0f", reportRunAddr, field.Name, value.Interface().(float64))
				// fmt.Sprintf("http://%s:%s/update/gauge/%s/%.0f", reportHost, reportPort, field.Name, value)
			} else if value.Kind() == reflect.Uint32 {
				urlGauge = fmt.Sprintf("http://%s/update/gauge/%s/%d", reportRunAddr, field.Name, value.Interface().(uint32))
			} else if value.Kind() == reflect.Uint64 {
				urlGauge = fmt.Sprintf("http://%s/update/gauge/%s/%d", reportRunAddr, field.Name, value.Interface().(uint64))
			}
			urlCounter := fmt.Sprintf("http://%s/update/counter/%s/%d", reportRunAddr, field.Name, metrics.PollCount)

			req, _ := http.NewRequest(http.MethodPost, urlGauge, nil)
			req.Header.Add("Content-Type", "text/plain")
			res, err := httpClient.Do(req)
			if err != nil {
				return err
			}
			res.Body.Close()
			//fmt.Println(res.Status)

			req, _ = http.NewRequest(http.MethodPost, urlCounter, nil)
			req.Header.Add("Content-Type", "text/plain")
			res, err = httpClient.Do(req)
			if err != nil {
				return err
			}
			res.Body.Close()
			//fmt.Println(res.Status)
		}
	}

	return nil
}

func main() {
	flags.ParseFlags()

	pollInterval = flags.FlagPollInterval
	reportInterval = flags.FlagReportInterval
	reportRunAddr = flags.FlagRunAddr

	pollSum := 0
	metricsStorage := metrics.New()

	for {
		runtime.ReadMemStats(&rtm)
		pollSum += pollInterval
		metricsStorage.RandomValueUpdate()
		metricsStorage.PollCountInc()
		time.Sleep(time.Duration(pollInterval) * time.Second)

		if pollSum >= reportInterval {
			err := sendMetrics(metricsStorage)
			if err != nil {
				panic(err)
			}
			pollSum = 0
			metricsStorage.PollCountDrop()
		}
	}
}
