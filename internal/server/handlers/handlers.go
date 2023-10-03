package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/models"
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
	status       int
	contEncoding string
	size         int
}

// проверяем, что клиент отправил серверу сжатые данные в формате gzip
func reqCheckGzipBody(r *http.Request) (io.ReadCloser, error) {
	contentEncoding := r.Header.Get("Content-Encoding")
	sendsGzip := strings.Contains(contentEncoding, "gzip")

	if sendsGzip {
		gzr, err := gzip.NewReader(r.Body)
		if err != nil {
			return r.Body, err
		}
		// меняем тело запроса на новое
		r.Body = gzr
		defer gzr.Close()
	}

	return r.Body, nil
}

// добавляем кастомную реализацию http.ResponseWriter
type loggingGzipResponseWriter struct {
	// встраиваем оригинальный http.ResponseWriter
	http.ResponseWriter
	responseData *responseData
	resCompress  bool // требуется ли сжимать ответ
}

func (r *loggingGzipResponseWriter) Write(b []byte) (int, error) {
	var (
		size int
		err  error
		gzw  *gzip.Writer
	)

	if r.resCompress {
		gzw, err = gzip.NewWriterLevel(r.ResponseWriter, gzip.BestSpeed)
		if err != nil {
			return -1, err
		}
		defer gzw.Close()

		r.responseData.contEncoding = "gzip" // сохраняем значение contEncoding

		size, err = gzw.Write(b)

	} else {
		// записываем ответ, используя оригинальный http.ResponseWriter
		size, err = r.ResponseWriter.Write(b)
	}

	r.responseData.size += size // захватываем размер

	return size, err
}

func (r *loggingGzipResponseWriter) WriteHeaderStatus(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func logHTTPResult(start time.Time, lw loggingGzipResponseWriter, r http.Request, optErr ...error) {
	err := errors.New("null")
	if len(optErr) > 0 {
		err = optErr[0]
	}

	log.Info().
		Str("URI", r.URL.Path).
		Str("Method", r.Method).
		Dur("duration", time.Since(start)).
		Msg("request")

	log.Info().
		Str("Status", strconv.Itoa(lw.responseData.status)).
		Str("Content-Length", strconv.Itoa(lw.responseData.size)).
		Err(err).
		Msg("response")
}

func UpdateMetric(reqJSON models.Metrics, repo storage.Storer) error {
	if reqJSON.MType == "gauge" {
		value := reqJSON.Value
		if value == nil {
			return fmt.Errorf("bad gauge value")
		}
		repo.SetGauge(reqJSON.ID, storage.Gauge(*value))
	} else if reqJSON.MType == "counter" {
		value := reqJSON.Delta
		if value == nil {
			return fmt.Errorf("bad counetr delta")
		}
		repo.UpdateCounter(reqJSON.ID, storage.Counter(*value))
	} else {
		return fmt.Errorf("bad metric type: %s", reqJSON.MType)
	}

	return nil
}

func getReqContEncoding(r *http.Request) bool {

	encodingSlice := r.Header.Values("Accept-Encoding")
	encodingsStr := strings.Join(encodingSlice, ",")
	encodings := strings.Split(encodingsStr, ",")

	for _, el := range encodings {
		if el == "gzip" {
			return true
		}
	}

	return false
}

func UpdateHandlerLong(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var (
		valCounter       storage.Counter
		valGauge         storage.Gauge
		ok               bool
		reqJSON, resJSON models.Metrics
	)
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingGzipResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
		resCompress:    false,
	}

	reqJSON.ID = chi.URLParam(r, "metricName")
	reqJSON.MType = chi.URLParam(r, "metricType")

	if reqJSON.MType == "counter" {
		counterVal, err := strconv.ParseInt(chi.URLParam(r, "metricVal"), 10, 64)
		if err != nil {
			lw.WriteHeaderStatus(http.StatusBadRequest)
			logHTTPResult(start, lw, *r, err)
			return
		}
		reqJSON.Delta = &counterVal
	} else if reqJSON.MType == "gauge" {
		gaugeVal, err := strconv.ParseFloat(chi.URLParam(r, "metricVal"), 64)
		if err != nil {
			lw.WriteHeaderStatus(http.StatusBadRequest)
			logHTTPResult(start, lw, *r, err)
			return
		}
		reqJSON.Value = &gaugeVal
	} else {
		err := fmt.Errorf("can not get val for %v from repo", reqJSON.MType)
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	lw.Header().Set("Content-Type", "application/json")
	lw.Header().Set("Date", time.Now().String())

	err := UpdateMetric(reqJSON, repo)
	if err != nil {
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	resJSON = reqJSON
	if resJSON.MType == "counter" {
		if valCounter, ok = repo.GetCounter(resJSON.ID); ok {
			valCounterI64 := int64(valCounter)
			resJSON.Delta = &valCounterI64
		}
	} else if resJSON.MType == "gauge" {
		if valGauge, ok = repo.GetGauge(resJSON.ID); ok {
			valGaugeF64 := float64(valGauge)
			resJSON.Value = &valGaugeF64
		}
	} else {
		err := fmt.Errorf("can not get val for %v from repo", reqJSON.ID)
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	if ok {
		lw.WriteHeaderStatus(http.StatusOK)
	} else {
		lw.WriteHeaderStatus(http.StatusNotFound)
	}

	fmt.Println("update_long-reqJSON", reqJSON.String())
	fmt.Println("update_long-resJSON", resJSON.String())

	logHTTPResult(start, lw, *r)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var (
		reqJSON, resJSON models.Metrics
		valCounter       storage.Counter
		valGauge         storage.Gauge
		ok               bool
		err              error
	)
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingGzipResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
		resCompress:    false,
	}

	r.Body, err = reqCheckGzipBody(r)
	if err != nil {
		lw.WriteHeaderStatus(http.StatusInternalServerError)
		logHTTPResult(start, lw, *r, err)
		return
	}

	if lw.resCompress = getReqContEncoding(r); lw.resCompress {
		lw.Header().Set("Content-Encoding", "gzip")
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqJSON); err != nil {
		lw.WriteHeaderStatus(http.StatusInternalServerError)
		logHTTPResult(start, lw, *r, err)
		return
	}

	lw.Header().Set("Content-Type", "application/json")
	lw.Header().Set("Date", time.Now().String())

	err = UpdateMetric(reqJSON, repo)
	if err != nil {
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	resJSON.ID = reqJSON.ID
	resJSON.MType = reqJSON.MType

	if resJSON.MType == "counter" {
		if valCounter, ok = repo.GetCounter(resJSON.ID); ok {
			valCounterI64 := int64(valCounter)
			resJSON.Delta = &valCounterI64
		}
	} else if resJSON.MType == "gauge" {
		if valGauge, ok = repo.GetGauge(resJSON.ID); ok {
			valGaugeF64 := float64(valGauge)
			resJSON.Value = &valGaugeF64
		}
	} else {
		err := fmt.Errorf("can not get val for %v from repo", resJSON.ID)
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	lw.WriteHeaderStatus(http.StatusOK)

	enc := json.NewEncoder(&lw)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resJSON); err != nil {
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	//fmt.Println("update-reqJSON", reqJSON.String())
	//fmt.Println("update-resJSON", resJSON.String())

	logHTTPResult(start, lw, *r)
}

func ValueHandlerLong(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var (
		valCounter       storage.Counter
		valGauge         storage.Gauge
		ok               bool
		data             string
		reqJSON, resJSON models.Metrics
	)
	start := time.Now()

	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingGzipResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
		resCompress:    false,
	}

	reqJSON.ID = chi.URLParam(r, "metricName")
	reqJSON.MType = chi.URLParam(r, "metricType")

	lw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	lw.Header().Set("Date", time.Now().String())

	resJSON = reqJSON
	if resJSON.MType == "counter" {
		if valCounter, ok = repo.GetCounter(resJSON.ID); ok {
			data = fmt.Sprintf("%d", valCounter)
		}
	} else if resJSON.MType == "gauge" {
		if valGauge, ok = repo.GetGauge(resJSON.ID); ok {
			data = fmt.Sprintf("%.3f", valGauge)
			data = strings.Trim(data, "0")
		}
	} else {
		err := fmt.Errorf("can not get val for %v from repo", reqJSON.ID)
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	if ok {
		lw.WriteHeaderStatus(http.StatusOK)
		lw.Write([]byte(data))
	} else {
		lw.WriteHeaderStatus(http.StatusNotFound)
	}

	logHTTPResult(start, lw, *r)
}

func ValueHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	var (
		valCounter       storage.Counter
		valGauge         storage.Gauge
		ok               bool
		err              error
		reqJSON, resJSON models.Metrics
	)
	start := time.Now()

	responseData := &responseData{
		status:       0,
		contEncoding: "",
		size:         0,
	}
	lw := loggingGzipResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
		resCompress:    false,
	}

	r.Body, err = reqCheckGzipBody(r)
	if err != nil {
		lw.WriteHeaderStatus(http.StatusInternalServerError)
		logHTTPResult(start, lw, *r, err)
		return
	}

	if lw.resCompress = getReqContEncoding(r); lw.resCompress {
		lw.Header().Set("Content-Encoding", "gzip")
	}

	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&reqJSON); err != nil {
		lw.WriteHeaderStatus(http.StatusInternalServerError)
		logHTTPResult(start, lw, *r, err)
		return
	}

	lw.Header().Set("Content-Type", "application/json")
	lw.Header().Set("Date", time.Now().String())

	resJSON.ID = reqJSON.ID
	resJSON.MType = reqJSON.MType

	if resJSON.MType == "counter" {
		if valCounter, ok = repo.GetCounter(resJSON.ID); ok {
			valCounterI64 := int64(valCounter)
			resJSON.Delta = &valCounterI64
		}
	} else if resJSON.MType == "gauge" {
		if valGauge, ok = repo.GetGauge(reqJSON.ID); ok {
			valGaugeF64 := float64(valGauge)
			resJSON.Value = &valGaugeF64
		}
	} else {
		err = fmt.Errorf("can not get val for %v from repo", resJSON.MType)
		lw.WriteHeaderStatus(http.StatusBadRequest)
		logHTTPResult(start, lw, *r, err)
		return
	}

	if ok {
		lw.WriteHeaderStatus(http.StatusOK)
		enc := json.NewEncoder(&lw)
		enc.SetIndent("", "  ")
		if err := enc.Encode(resJSON); err != nil {
			lw.WriteHeaderStatus(http.StatusBadRequest)
			logHTTPResult(start, lw, *r, err)
			return
		}
	} else {
		err = fmt.Errorf("can not get val for <%v>, type <%v> from repo", reqJSON.ID, reqJSON.MType)
		lw.WriteHeaderStatus(http.StatusNotFound)
	}

	//fmt.Println("value-reqJSON", reqJSON.String())
	//fmt.Println("value-resJSON", resJSON.String())

	logHTTPResult(start, lw, *r, err)
}

func RootHandler(w http.ResponseWriter, r *http.Request, repo storage.Storer) {
	start := time.Now()

	responseData := &responseData{
		status:       0,
		contEncoding: "",
		size:         0,
	}
	lw := loggingGzipResponseWriter{
		ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
		responseData:   responseData,
		resCompress:    false,
	}

	if lw.resCompress = getReqContEncoding(r); lw.resCompress {
		lw.Header().Set("Content-Encoding", "gzip")
	}

	lw.Header().Set("Content-Type", "text/html")
	lw.Header().Set("Date", time.Now().String())

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

	data := []byte(WebPage)

	lw.WriteHeaderStatus(http.StatusOK)

	if _, err := lw.Write(data); err != nil {
		panic(err)
	}

	//if _, err := io.WriteString(&lw, WebPage); err != nil {
	//	panic(err)
	//}

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
