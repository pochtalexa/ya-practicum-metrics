package handlers

import (
	"fmt"
	"net/http"
	"time"
	"strings"
	"errors"
	"slices"
	"strconv"

	"github.com/pochtalexa/ya-practicum-metrics/internal/server/storage"
)


var MemStorage = storage.NewMemStore()


func checkMethod(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {        
        w.WriteHeader(http.StatusMethodNotAllowed)
        return errors.New("bad Method")
    }
	return nil
}

func urlParse(w http.ResponseWriter, url string) (map[string]string, error) {
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
    if err != nil {		
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("bad metricVal")

    }

	w.WriteHeader(http.StatusOK)
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
	var err error

    if err = checkMethod(w, r); err != nil {
		return
	}
	
	CurMetric, err := urlParse(w, r.URL.Path)
	if err != nil {		
		return
	}

	fmt.Println(CurMetric)

	err = UpdateMetric(CurMetric)
	if err != nil {		
		return
	}
	
	fmt.Println(MemStorage)
}