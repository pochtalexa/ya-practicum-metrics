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
	//var respMetric models.Metrics
	urlMetric := fmt.Sprintf("http://%s/update/", reportRunAddr)

	for _, el := range CashMetrics.CashMetrics {
		respMetric := models.Metrics{}

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

		dec := json.NewDecoder(res.Body)
		if err := dec.Decode(&respMetric); err != nil {
			log.Info().Err(err).Msg("decode body error")
			return err
		}
		res.Body.Close()

		log.Info().Str("status", res.Status).Msg(fmt.Sprintln("respMetric:", respMetric.String()))
	}

	return nil
}

func GetRoot(httpClient http.Client, reportRunAddr string) error {
	urlMetric := fmt.Sprintf("http://%s/", reportRunAddr)

	req, _ := http.NewRequest(http.MethodGet, urlMetric, nil)
	req.Header.Add("Content-Type", "text/plain; charset=utf-8")

	res, err := http.Get(urlMetric)
	//res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	log.Info().Str("status", res.Status).Msg("root page")
	return nil
}
