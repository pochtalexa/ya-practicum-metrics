package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/models"
	"github.com/rs/zerolog/log"
	"net/http"
)

func CollectMetrics(metrics *RuntimeMetrics) (CashMetrics, error) {
	var (
		CashMetrics   CashMetrics
		gaugeMetric   models.Metrics
		counterMetric models.Metrics
	)

	for _, mName := range metrics.GetGaugeName() {
		gaugeMetric.ID = mName
		gaugeMetric.MType = "gauge"

		gaugeMetricTemp, err := metrics.GetGaugeValue(mName)
		if err != nil {
			return CashMetrics, err
		}
		gaugeMetric.Value = &gaugeMetricTemp

		CashMetrics.CashMetrics = append(CashMetrics.CashMetrics, gaugeMetric)
	}

	counterMetric.ID = "PollCount"
	counterMetric.MType = "counter"
	counterMetricTemp := int64(metrics.PollCount)
	counterMetric.Delta = &counterMetricTemp

	CashMetrics.CashMetrics = append(CashMetrics.CashMetrics, counterMetric)

	return CashMetrics, nil
}

func SendMetric(CashMetrics CashMetrics, httpClient http.Client, reportRunAddr string) error {
	urlMetric := fmt.Sprintf("http://%s/update/", reportRunAddr)

	for _, el := range CashMetrics.CashMetrics {
		respMetric := models.Metrics{}

		reqBody, err := json.Marshal(el)
		if err != nil {
			panic(err)
		}
		log.Info().Str("reqBody", string(reqBody)).Msg("Marshal result")

		req, _ := http.NewRequest(http.MethodPost, urlMetric, bytes.NewReader(reqBody))
		req.Header.Add("Content-Type", "application/json")
		res, err := httpClient.Do(req)
		if err != nil {
			log.Info().Err(err).Msg("SendMetric error")
			continue
		}
		defer res.Body.Close()

		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(&respMetric); err != nil {
			log.Info().Err(err).Msg("decode body error")
			continue
		}

		log.Info().Str("status", res.Status).Msg(fmt.Sprintln("respMetric:", respMetric.String()))
	}

	return nil
}
