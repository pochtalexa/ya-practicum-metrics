package main

import (
	"fmt"
	patronhttp "github.com/beatlabs/patron/client/http"
	"net/http"
	"reflect"
	"runtime"
	//"slices"
	"golang.org/x/exp/slices"
	"time"

	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
)

var rtm runtime.MemStats

const (
	pollInterval   = 2
	reportInterval = 10
	reportHost     = "127.0.0.1"
	reportPort     = "8080"
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
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%.0f", reportHost, reportPort, field.Name, value.Interface().(float64))
				// fmt.Sprintf("http://%s:%s/update/gauge/%s/%.0f", reportHost, reportPort, field.Name, value)
			} else if value.Kind() == reflect.Uint32 {
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%d", reportHost, reportPort, field.Name, value.Interface().(uint32))
			} else if value.Kind() == reflect.Uint64 {
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%d", reportHost, reportPort, field.Name, value.Interface().(uint64))
			}
			urlCounter := fmt.Sprintf("http://%s:%s/update/counter/%s/%d", reportHost, reportPort, field.Name, metrics.PollCount)

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
	pollSum := 0
	metricsStorage := metrics.New()

	for {
		runtime.ReadMemStats(&rtm)
		pollSum += pollInterval
		metricsStorage.RandomValueUpdate()
		metricsStorage.PollCountInc()
		time.Sleep(pollInterval * time.Second)

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
