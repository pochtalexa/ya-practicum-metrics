package handlers

import (
	"fmt"
	"net/http"
	"time"
	"strings"
	"errors"
	"slices"
	"strconv"	
)



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


func UpdateHandler(w http.ResponseWriter, r *http.Request) {
    if err := checkMethod(w, r); err != nil {
		return
	}
	
	CurMetric, err := urlParse(w, r.URL.Path)
	if err != nil {		
		return
	}

	fmt.Println(CurMetric)
}