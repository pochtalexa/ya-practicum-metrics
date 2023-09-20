package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
)

var CurMetric = make(map[string]string)

func UpdateMetric(CurMetric map[string]string, MemStorage storage.Store) error {
	if CurMetric["metricType"] == "gauge" {
		value, _ := strconv.ParseFloat(CurMetric["metricVal"], 64)
		MemStorage.SetGauge(CurMetric["metricName"], storage.Gauge(value))
	} else {
		value, _ := strconv.ParseInt(CurMetric["metricVal"], 10, 64)
		MemStorage.UpdateCounter(CurMetric["metricName"], storage.Counter(value))
	}
	return nil
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, MemStorage storage.Store) {

	CurMetric["metricType"] = chi.URLParam(r, "metricType")
	CurMetric["metricName"] = chi.URLParam(r, "metricName")
	CurMetric["metricVal"] = chi.URLParam(r, "metricVal")

	w.WriteHeader(http.StatusOK)

	err := UpdateMetric(CurMetric, MemStorage)
	if err != nil {
		return
	}
}

func ValueHandler(w http.ResponseWriter, r *http.Request, MemStorage storage.Store) {
	var (
		valCounter storage.Counter
		valGauge   storage.Gauge
		ok         bool
		data       string
	)

	CurMetric["metricType"] = chi.URLParam(r, "metricType")
	CurMetric["metricName"] = chi.URLParam(r, "metricName")
	CurMetric["metricVal"] = chi.URLParam(r, "metricVal")

	if CurMetric["metricType"] == "counter" {
		if valCounter, ok = MemStorage.GetCounter(CurMetric["metricName"]); ok {
			data = fmt.Sprintf("%d", valCounter)
		}
	} else {
		if valGauge, ok = MemStorage.GetGauge(CurMetric["metricName"]); ok {
			data = fmt.Sprintf("%.3f", valGauge)
			data = strings.Trim(data, "0")
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().String())

	if ok {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(data))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func RootHandler(w http.ResponseWriter, r *http.Request, MemStorage storage.Store) {

	WebPage1, _ := MemStorage.String("gauges")
	WebPage2, _ := MemStorage.String("counters")

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

	if _, err := io.WriteString(w, WebPage); err != nil {
		panic(err)
	}
}
