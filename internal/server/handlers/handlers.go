package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
)

// структура для хранения сведений об ответе
type responseData struct {
	status int
	size   int
}

// добавляем кастомную реализацию http.ResponseWriter
type loggingResponseWriter struct {
	// встраиваем оригинальный http.ResponseWriter
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func logHTTPResult(start time.Time, lw loggingResponseWriter, r http.Request) {
	log.Info().
		Str("URI", r.URL.Path).
		Str("Method", r.Method).
		Dur("duration", time.Now().Sub(start)).
		Msg("request")

	log.Info().
		Str("Status", strconv.Itoa(lw.responseData.status)).
		Str("Content-Length", strconv.Itoa(lw.responseData.size)).
		Msg("response")
}

func UpdateMetric(CurMetric map[string]string, repo storage.Storer) error {
	if CurMetric["metricType"] == "gauge" {
		value, err := strconv.ParseFloat(CurMetric["metricVal"], 64)
		if err != nil {
			return fmt.Errorf("bad gauge val: %s", CurMetric["metricVal"])
		}
		repo.SetGauge(CurMetric["metricName"], storage.Gauge(value))
	} else if CurMetric["metricType"] == "counter" {
		value, err := strconv.ParseInt(CurMetric["metricVal"], 10, 64)
		if err != nil {
			return fmt.Errorf("bad counter val: %s", CurMetric["metricVal"])
		}
		repo.UpdateCounter(CurMetric["metricName"], storage.Counter(value))
	} else {
		return fmt.Errorf("bad metric type val: %s", CurMetric["metricType"])
	}

	return nil
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var CurMetric = make(map[string]string)
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
	}

	CurMetric["metricType"] = chi.URLParam(r, "metricType")
	CurMetric["metricName"] = chi.URLParam(r, "metricName")
	CurMetric["metricVal"] = chi.URLParam(r, "metricVal")

	lw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	lw.Header().Set("Date", time.Now().String())

	err := UpdateMetric(CurMetric, repo)
	if err != nil {
		lw.WriteHeader(http.StatusBadRequest)
		logHTTPResult(start, lw, *r)
		return
	}

	lw.WriteHeader(http.StatusOK)

	logHTTPResult(start, lw, *r)
}

func ValueHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var CurMetric = make(map[string]string)
	var (
		valCounter storage.Counter
		valGauge   storage.Gauge
		ok         bool
		data       string
	)
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
	}

	CurMetric["metricType"] = chi.URLParam(r, "metricType")
	CurMetric["metricName"] = chi.URLParam(r, "metricName")
	CurMetric["metricVal"] = chi.URLParam(r, "metricVal")

	if CurMetric["metricType"] == "counter" {
		if valCounter, ok = repo.GetCounter(CurMetric["metricName"]); ok {
			data = fmt.Sprintf("%d", valCounter)
		}
	} else {
		if valGauge, ok = repo.GetGauge(CurMetric["metricName"]); ok {
			data = fmt.Sprintf("%.3f", valGauge)
			data = strings.Trim(data, "0")
		}
	}

	lw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	lw.Header().Set("Date", time.Now().String())

	if ok {
		lw.WriteHeader(http.StatusOK)
		lw.Write([]byte(data))
	} else {
		lw.WriteHeader(http.StatusNotFound)
	}

	logHTTPResult(start, lw, *r)
}

func RootHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
	}

	WebPage1, _ := gauges2String(repo.GetGauges())
	WebPage2, _ := сounters2String(repo.GetCounters())
	WebPage := fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Document</title>
	</head>
	<body>
		<h3>Metric values</h3>
		<h5>gauges</h5>
		<p> %s </p>
		<p> </p>
		<h5>counters</h5>
		<p> %s </p>
	</body>
	</html>`, WebPage1, WebPage2)

	lw.WriteHeader(http.StatusOK)

	if _, err := io.WriteString(&lw, WebPage); err != nil {
		panic(err)
	}

	logHTTPResult(start, lw, *r)
}

func сounters2String(mapCounters map[string]storage.Counter) (string, error) {
	var storeList []string

	for k, v := range mapCounters {
		storeList = append(storeList, k+":"+fmt.Sprintf("%d", v))
	}

	return strings.Join(storeList, ","), nil
}

func gauges2String(mapGauges map[string]storage.Gauge) (string, error) {
	var storeList []string

	for k, v := range mapGauges {
		storeList = append(storeList, k+":"+fmt.Sprintf("%.3f", v))
	}

	return strings.Join(storeList, ","), nil
}
