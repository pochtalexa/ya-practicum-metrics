package metrics

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/models"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-retry"
	"net"
	"net/http"
	"strings"
	"time"
)

type hashSHA256Struct struct {
	data string
	err  error
}

var hashSHA256 = &hashSHA256Struct{}

func signReqBody(body []byte) (string, error) {
	h := hmac.New(sha256.New, []byte(flags.FlagHashKey))
	h.Write(body)
	dst := h.Sum(nil)

	return hex.EncodeToString(dst), nil
}

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

func SendMetricBatch(CashMetrics CashMetrics, httpClient http.Client, reportRunAddr string) error {
	var netErr net.Error
	urlMetric := fmt.Sprintf("http://%s/updates/", reportRunAddr)
	ctx := context.Background()
	b := retry.NewFibonacci(1 * time.Second)

	type responseBody struct {
		Description string `json:"description"` // имя метрики
	}

	resBody := responseBody{}

	reqBody, err := json.Marshal(CashMetrics.CashMetrics)
	if err != nil {
		return fmt.Errorf("marshal Batch error, %w", err)
	}
	log.Info().Str("reqBody", string(reqBody)).Msg("Marshal Batch result")

	if flags.UseHashKey {
		hashSHA256.data, hashSHA256.err = signReqBody(reqBody)
		if hashSHA256.err != nil {
			log.Info().Err(err).Msg("can not signReqBody")
		}
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	gzipWriter.Write(reqBody)
	gzipWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, urlMetric, &buf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Encoding", "gzip")

	if flags.UseHashKey && hashSHA256.err == nil {
		req.Header.Add("HashSHA256", hashSHA256.data)
	}

	err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
		res, err := httpClient.Do(req)
		if err != nil {
			if errors.As(err, &netErr) ||
				netErr.Timeout() ||
				strings.Contains(err.Error(), "EOF") ||
				strings.Contains(err.Error(), "connection reset by peer") {

				return retry.RetryableError(err)
			}
			return err
		}
		defer res.Body.Close()

		log.Info().Str("status", res.Status).Msg(fmt.Sprintln("resBody Batch:", resBody.Description))
		return nil
	})
	if err != nil {
		return fmt.Errorf("SendMetric error, %w", err)
	}

	return nil
}

func SendMetric(CashMetrics CashMetrics, httpClient http.Client, reportRunAddr string) error {
	var netErr net.Error
	urlMetric := fmt.Sprintf("http://%s/update/", reportRunAddr)

	for _, el := range CashMetrics.CashMetrics {
		ctx := context.Background()
		b := retry.NewFibonacci(1 * time.Second)
		respMetric := models.Metrics{}

		reqBody, err := json.Marshal(el)
		if err != nil {
			return fmt.Errorf("marshal error, %w", err)
		}
		log.Info().Str("reqBody", string(reqBody)).Msg("Marshal result")

		if flags.UseHashKey {
			hashSHA256.data, hashSHA256.err = signReqBody(reqBody)
			if hashSHA256.err != nil {
				log.Info().Err(err).Msg("can not signReqBody")
			}
		}

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		gzipWriter.Write(reqBody)
		gzipWriter.Close()

		req, _ := http.NewRequest(http.MethodPost, urlMetric, &buf)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Encoding", "gzip")

		if flags.UseHashKey && hashSHA256.err != nil {
			req.Header.Add("HashSHA256", hashSHA256.data)
		}

		err = retry.Do(ctx, retry.WithMaxRetries(3, b), func(ctx context.Context) error {
			res, err := httpClient.Do(req)
			if err != nil {
				if errors.As(err, &netErr) ||
					netErr.Timeout() ||
					strings.Contains(err.Error(), "EOF") ||
					strings.Contains(err.Error(), "connection reset by peer") {

					return retry.RetryableError(err)
				}
				return err
			}
			defer res.Body.Close()

			dec := json.NewDecoder(res.Body)
			if err := dec.Decode(&respMetric); err != nil {
				return fmt.Errorf("decode body error, %w", err)
			}

			log.Info().Str("status", res.Status).Msg(fmt.Sprintln("respMetric:", respMetric.String()))
			return nil
		})

		if err != nil {
			log.Info().Err(err).Msg("send metric error")
		}
	}

	return nil
}
