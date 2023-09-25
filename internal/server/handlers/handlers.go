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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().String())

	CurMetric["metricType"] = chi.URLParam(r, "metricType")
	CurMetric["metricName"] = chi.URLParam(r, "metricName")
	CurMetric["metricVal"] = chi.URLParam(r, "metricVal")

	err := UpdateMetric(CurMetric, repo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ValueHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var CurMetric = make(map[string]string)

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
		if valCounter, ok = repo.GetCounter(CurMetric["metricName"]); ok {
			data = fmt.Sprintf("%d", valCounter)
		}
	} else {
		if valGauge, ok = repo.GetGauge(CurMetric["metricName"]); ok {
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
	log.Info().Str("URI", r.URL).Msg(r.URL)

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

	if _, err := io.WriteString(w, WebPage); err != nil {
		panic(err)
	}
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
