package metrics

import (
	"fmt"
	patronhttp "github.com/beatlabs/patron/client/http"
	"net/http"
	"strconv"
)

var err error

func CollectMetrics(metrics *RuntimeMetrics) (CashMetrics, error) {
	var CashMetrics CashMetrics
	var gaugeMetric GaugeMetric
	var counterMetric CounterMetric

	for _, mName := range metrics.GetMericsName() {
		gaugeMetric.Name = mName
		gaugeMetric.Value, err = metrics.GetDataValue(mName)
		if err != nil {
			return CashMetrics, err
		}
		gaugeMetric.ValueStr = strconv.FormatFloat(float64(gaugeMetric.Value), 'E', -1, 64)

		counterMetric.Name = mName
		counterMetric.Value = metrics.PollCount
		counterMetric.ValueStr = strconv.FormatUint(uint64(counterMetric.Value), 10)

		CashMetrics.GaugeMetrics = append(CashMetrics.GaugeMetrics, gaugeMetric)
		CashMetrics.CounterMetric = append(CashMetrics.CounterMetric, counterMetric)
	}

	return CashMetrics, nil
}

func SendMetric(CashMetrics CashMetrics, httpClient patronhttp.Client, reportRunAddr string) error {

	for _, el := range CashMetrics.GaugeMetrics {
		urlGauge := fmt.Sprintf("http://%s/update/gauge/%s/%s", reportRunAddr, el.Name, el.ValueStr)

		req, _ := http.NewRequest(http.MethodPost, urlGauge, nil)
		req.Header.Add("Content-Type", "text/plain")
		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
	}

	for _, el := range CashMetrics.CounterMetric {
		urlCounter := fmt.Sprintf("http://%s/update/counter/%s/%s", reportRunAddr, el.Name, el.ValueStr)

		req, _ := http.NewRequest(http.MethodPost, urlCounter, nil)
		req.Header.Add("Content-Type", "text/plain")
		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
	}

	return nil
}
