package send

import (
	"fmt"
	patronhttp "github.com/beatlabs/patron/client/http"
	"github.com/pochtalexa/ya-practicum-metrics/internal/agent/metrics"
	"net/http"
)

func Metrics(metrics *metrics.RuntimeMetrics, reportRunAddr string) error {
	var urlGauge string

	//httpClient := &http.Client{}
	httpClient, err := patronhttp.New()
	if err != nil {
		return err
	}

	for _, mName := range metrics.GaugesName {
		mValue, err := metrics.GetGaugeValue(mName)
		if err != nil {
			return err
		}

		urlGauge = fmt.Sprintf("http://%s/update/gauge/%s/%f", reportRunAddr, mName, mValue)
		urlCounter := fmt.Sprintf("http://%s/update/counter/%s/%d", reportRunAddr, mName, metrics.PollCount)

		req, _ := http.NewRequest(http.MethodPost, urlGauge, nil)
		req.Header.Add("Content-Type", "text/plain")
		res, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()

		req, _ = http.NewRequest(http.MethodPost, urlCounter, nil)
		req.Header.Add("Content-Type", "text/plain")
		res, err = httpClient.Do(req)
		if err != nil {
			return err
		}
		res.Body.Close()
	}

	return nil
}
