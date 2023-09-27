package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func CollectMetrics(metrics *RuntimeMetrics) (CashMetrics, error) {
	var (
		CashMetrics   CashMetrics
		gaugeMetric   Metrics
		counterMetric Metrics
	)

	for _, mName := range metrics.GetMericsName() {
		gaugeMetric.ID = mName
		gaugeMetric.MType = "gauge"

		gaugeMetricTemp, err := metrics.GetDataValue(mName)
		if err != nil {
			return CashMetrics, err
		}
		gaugeMetric.Value = &gaugeMetricTemp

		counterMetric.ID = mName
		counterMetric.MType = "counter"
		counterMetricTemp := int64(metrics.PollCount)
		counterMetric.Delta = &counterMetricTemp

		CashMetrics.CashMetrics = append(CashMetrics.CashMetrics, gaugeMetric)
		CashMetrics.CashMetrics = append(CashMetrics.CashMetrics, counterMetric)
	}

	return CashMetrics, nil
}

func SendMetric(CashMetrics CashMetrics, httpClient http.Client, reportRunAddr string) error {
	urlMetric := fmt.Sprintf("http://%s/update/", reportRunAddr)

	for _, el := range CashMetrics.CashMetrics {

		reqBody, err := json.Marshal(el)
		if err != nil {
			panic(err)
		}
		req, _ := http.NewRequest(http.MethodPost, urlMetric, bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")
		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
	}

	return nil
}
