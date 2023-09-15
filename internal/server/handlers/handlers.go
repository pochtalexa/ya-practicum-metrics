package handlers

import (
	"fmt"
	"net/http"
	"time"
	"strings"
	"errors"
	"slices"
	"strconv"
	"io"
	
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
)


var MemStorage = storage.NewMemStore()

func checkMethodPost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {        
        w.WriteHeader(http.StatusMethodNotAllowed)
        return errors.New("bad Method")
    }
	return nil
}

func checkMethodGet(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {        
        w.WriteHeader(http.StatusMethodNotAllowed)
        return errors.New("bad Method")
    }
	return nil
}

func urlParse(w http.ResponseWriter, url string, action string) (map[string]string, error) {
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().String())
	
	MetricTypes := []string{"gauge", "counter"}
	CurMetric := make(map[string]string)

	urlParts := strings.Split(url, "/")

	for k, v := range urlParts {		
		switch k {			
			case 2: CurMetric["metricType"] = v
			case 3: CurMetric["metricName"] = v
			case 4: CurMetric["metricVal"] = v
		}
	}	

	if CurMetric["metricName"] == "" {		
		w.WriteHeader(http.StatusNotFound)
		return nil, errors.New("bad metricName")
	}
	
	if !slices.Contains(MetricTypes, CurMetric["metricType"]) {		
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("bad metricType")
	}

	_, err := strconv.Atoi(CurMetric["metricVal"])
    if err != nil && action == "update" {		
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("bad metricVal")

    }
	
	return CurMetric, nil
}

func UpdateMetric(CurMetric map[string]string) (error) {
	if CurMetric["metricType"] == "gauge" {
		value, _  := strconv.ParseInt(CurMetric["metricVal"], 10, 64)
		MemStorage.SetGauge(CurMetric["metricName"], storage.Gauge(value))		
	} else {
		value, _  := strconv.ParseInt(CurMetric["metricVal"], 10, 64)
		MemStorage.UpdateCounter(CurMetric["metricName"], storage.Counter(value))
	}	
	return nil
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

    if err := checkMethodPost(w, r); err != nil {
		return
	}
	
	CurMetric, err := urlParse(w, r.URL.Path, "upadte")
	if err != nil {		
		return
	}
	fmt.Println(CurMetric)

	w.WriteHeader(http.StatusOK)

	err = UpdateMetric(CurMetric)
	if err != nil {		
		return
	}

	fmt.Println(MemStorage)
}

func ValueHandler(w http.ResponseWriter, r *http.Request) {
	var (
		valCounter storage.Counter
		valGauge storage.Gauge
		ok bool
		data string
	)
	

	if err := checkMethodGet(w, r); err != nil {
		return
	}

	CurMetric, err := urlParse(w, r.URL.Path, "value")
	if err != nil {		
		return
	}


	if CurMetric["metricType"] == "counter" {
		if valCounter, ok = MemStorage.GetCounter(CurMetric["metricName"]); ok {
			data = CurMetric["metricType"] + ":" + CurMetric["metricName"] + ":" + fmt.Sprintf("%d", valCounter)
		} 		
	} else {
		if valGauge, ok = MemStorage.GetGauge(CurMetric["metricName"]); ok {
			data = CurMetric["metricType"] + ":" + CurMetric["metricName"] + ":" + fmt.Sprintf("%f", valGauge)
		}
	}		

	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().String())

	if ok {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(data))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}    
}

func RootHandler(w http.ResponseWriter, r *http.Request) {	
	if err := checkMethodGet(w, r); err != nil {
		return
	}	

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

	io.WriteString(w, WebPage)
}
