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
		value, err := strconv.ParseFloat(CurMetric["metricVal"], 64)
		if err != nil {
			return fmt.Errorf("bad gauge val: %s", CurMetric["metricVal"])
		}
		MemStorage.SetGauge(CurMetric["metricName"], storage.Gauge(value))
	} else if CurMetric["metricType"] == "counter" {
		value, err := strconv.ParseInt(CurMetric["metricVal"], 10, 64)
		if err != nil {
			return fmt.Errorf("bad counter val: %s", CurMetric["metricVal"])
		}
		MemStorage.UpdateCounter(CurMetric["metricName"], storage.Counter(value))
	} else {
		return fmt.Errorf("bad metric type val: %s", CurMetric["metricType"])
	}

	return nil
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, MemStorage storage.Store) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().String())

	CurMetric["metricType"] = chi.URLParam(r, "metricType")
	CurMetric["metricName"] = chi.URLParam(r, "metricName")
	CurMetric["metricVal"] = chi.URLParam(r, "metricVal")

	err := UpdateMetric(CurMetric, MemStorage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
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

func RootHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {

	WebPage1, _ := repo.String("gauges")
	WebPage2, _ := repo.String("counters")

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
