package metrics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/models"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"strconv"
	"time"
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

func httpClientDoRetry(httpClient http.Client, req *http.Request) (*http.Response, error) {
	var (
		res *http.Response
		err error
	)

	waiteIntervals := []int64{0, 1, 3, 5}
	waiteIntervalsLen := len(waiteIntervals) - 1
	for k, v := range waiteIntervals {
		var netErr net.Error

		time.Sleep(time.Second * time.Duration(v))
		res, err = httpClient.Do(req)
		if errors.As(err, &netErr) && k != waiteIntervalsLen {
			log.Info().
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("send metric attempt error")
			continue
		} else if errors.As(err, &netErr) && k == waiteIntervalsLen {
			log.Info().
				Str("attempt", strconv.FormatInt(int64(k), 10)).
				Msg("send metric attempt error")
			return nil, err
		} else if err != nil {
			return nil, err
		}
		break
	}

	return res, nil
}

func SendMetricBatch(CashMetrics CashMetrics, httpClient http.Client, reportRunAddr string) error {
	type responseBody struct {
		Description string `json:"description"` // имя метрики
	}
	var (
		res    *http.Response
		netErr net.Error
	)

	resBody := responseBody{}
	urlMetric := fmt.Sprintf("http://%s/updates/", reportRunAddr)

	reqBody, err := json.Marshal(CashMetrics.CashMetrics)
	if err != nil {
		panic(err)
	}
	log.Info().Str("reqBody", string(reqBody)).Msg("Marshal Batch result")

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	gzipWriter.Write(reqBody)
	gzipWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, urlMetric, &buf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	res, err = httpClientDoRetry(httpClient, req)
	if errors.As(err, &netErr) {
		log.Info().Err(err).Msg("SendMetric error")
		return nil
	} else if err != nil {
		log.Info().Err(err).Msg("error")
		return err
	}
	defer res.Body.Close()

	//dec := json.NewDecoder(res.Body)
	//if err := dec.Decode(&resBody); err != nil {
	//	log.Info().Err(err).Msg("decode body error")
	//	return err
	//}

	log.Info().Str("status", res.Status).Msg(fmt.Sprintln("resBody:", resBody.Description))

	return nil
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

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		gzipWriter.Write(reqBody)
		gzipWriter.Close()

		req, _ := http.NewRequest(http.MethodPost, urlMetric, &buf)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Encoding", "gzip")

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
