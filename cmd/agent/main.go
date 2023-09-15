package main

import (
	"fmt"	
	"reflect"
	"runtime"
	"slices"
	"time"
	"net/http"

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
	client := &http.Client{}
	
	fields := reflect.TypeOf(rtm)
	values := reflect.ValueOf(rtm)
	num := fields.NumField()
	
	for i := 0; i < num; i++ {
    	field := fields.Field(i)
    	value := values.Field(i)
		
		if slices.Contains(metrics.MetricsName, field.Name) {		
    		fmt.Println("Type:", field.Type, ",", field.Name, "=", value)

			if value.Kind() == reflect.Float64 {										
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%.0f", reportHost, reportPort, field.Name, value.Interface().(float64))
				// urlGauge = strings.Join([]string{"111", "222"}, "")
				// fmt.Sprintf("http://%s:%s/update/gauge/%s/%.0f", reportHost, reportPort, field.Name, value)
			} else if value.Kind() == reflect.Uint32 {
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%d", reportHost, reportPort, field.Name, value.Interface().(uint32))
			} else if value.Kind() == reflect.Uint64 {
				urlGauge = fmt.Sprintf("http://%s:%s/update/gauge/%s/%d", reportHost, reportPort, field.Name, value.Interface().(uint64))
			}
			urlCounter := fmt.Sprintf("http://%s:%s/update/counter/%s/%d", reportHost, reportPort, field.Name, metrics.PollCount)
						
			req, _ := http.NewRequest(http.MethodPost, urlGauge, nil)
			req.Header.Add("Content-Type", "text/plain")			
			res, err := client.Do(req)
			if err != nil {
				return err
			}
			res.Body.Close()
			fmt.Println(res.Status)

			req, _ = http.NewRequest(http.MethodPost, urlCounter, nil)
			req.Header.Add("Content-Type", "text/plain")			
			res, err = client.Do(req)
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
				panic(err)
			}			
		}
	}
}
