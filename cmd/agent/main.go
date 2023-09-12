package main

import (
	// "errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"slices"
	"time"

	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
)


var rtm runtime.MemStats

const (
	pollInterval = 2
	reportInterval = 10
	reportHost = "127.0.0.1"
	reportPort = "8080"
)

func sendMetrics(metrics *metrics.RuntimeMetrics) (error) {
	var urlGauge string
	
	fields := reflect.TypeOf(rtm)
	values := reflect.ValueOf(rtm)
	num := fields.NumField()
	
	for i := 0; i < num; i++ {
    	field := fields.Field(i)
    	value := values.Field(i)
		
		if slices.Contains(metrics.MetricsName, field.Name) {		
    		fmt.Println("Type:", field.Type, ",", field.Name, "=", value)

			if value.Kind() == reflect.Float64 {										
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%.0f", reportHost, reportPort, field.Name, value)
			} else {
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%d", reportHost, reportPort, field.Name, value)
			}
			urlCounter := fmt.Sprintf("http://%s:%s/update/counter/%s/%d", reportHost, reportPort, field.Name, metrics.PollCount)

			res, err := http.Post(urlGauge, "text/plain", nil)
			if err != nil {
				return err
			}
			res.Body.Close()
			fmt.Println(res.Status)

			res, err = http.Post(urlCounter, "text/plain", nil)
			if err != nil {
				return err
			}
			res.Body.Close()
			fmt.Println(res.Status)
		}
	}
	return nil
}


func main() {
	pollSumm := 0
	metricsStorage := metrics.New()

	for {		
		runtime.ReadMemStats(&rtm)
		pollSumm += pollInterval
		metricsStorage.RandomValueUpdate()
		metricsStorage.PollCountInc()
		time.Sleep(pollInterval * time.Second)

		if pollSumm >=  reportInterval {
			pollSumm = 0
			err := sendMetrics(metricsStorage)
			if err != nil {
				fmt.Println(err)
			}			
		}
	}
}
