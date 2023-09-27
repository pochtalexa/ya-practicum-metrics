package metrics

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
)

type Gauge float64
type Counter int64

type RuntimeMetrics struct {
	Data        runtime.MemStats
	MetricsName []string
	Counters    map[string]Counter
	PollCount   Counter
	RandomValue Gauge
}

//type GaugeMetric struct {
//	Name     string `json:"name"`
//	Value    Gauge  `json:"value"`
//	ValueStr string `json:"value_str"`
//}
//
//type CounterMetric struct {
//	Name     string  `json:"name"`
//	Value    Counter `json:"value"`
//	ValueStr string  `json:"value_str"`
//}

//type CashMetrics struct {
//	GaugeMetrics  []GaugeMetric
//	CounterMetric []CounterMetric
//}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type CashMetrics struct {
	CashMetrics []Metrics
}

func New() *RuntimeMetrics {
	return &RuntimeMetrics{
		MetricsName: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
			"HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys",
			"Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"},
		Counters: make(map[string]Counter),
	}
}

func (el *RuntimeMetrics) UpdateCounter(name string, value Counter) {
	el.Counters[name] = value
}

func (el *RuntimeMetrics) PollCountInc() {
	el.PollCount++
}

func (el *RuntimeMetrics) PollCountDrop() {
	el.PollCount = 0
}

func (el *RuntimeMetrics) RandomValueUpdate() {
	el.RandomValue = Gauge(rand.Float64() * math.Pow(10, 6))
}

func (el *RuntimeMetrics) GetMericsName() []string {
	return el.MetricsName
}

func (el *RuntimeMetrics) UpdateMetrics() {
	runtime.ReadMemStats(&el.Data)
	el.RandomValueUpdate()
	el.PollCountInc()
}

func (el *RuntimeMetrics) GetDataValue(name string) (float64, error) {
	var result float64

	switch name {
	case "Alloc":
		result = float64(el.Data.Alloc)
	case "BuckHashSys":
		result = float64(el.Data.BuckHashSys)
	case "Frees":
		result = float64(el.Data.Frees)
	case "GCSys":
		result = float64(el.Data.GCSys)
	case "HeapAlloc":
		result = float64(el.Data.HeapAlloc)
	case "HeapIdle":
		result = float64(el.Data.HeapIdle)
	case "HeapInuse":
		result = float64(el.Data.HeapInuse)
	case "HeapObjects":
		result = float64(el.Data.HeapObjects)
	case "HeapReleased":
		result = float64(el.Data.HeapReleased)
	case "HeapSys":
		result = float64(el.Data.HeapSys)
	case "LastGC":
		result = float64(el.Data.LastGC)
	case "Lookups":
		result = float64(el.Data.Lookups)
	case "MCacheInuse":
		result = float64(el.Data.MCacheInuse)
	case "MSpanSys":
		result = float64(el.Data.MSpanSys)
	case "MSpanInuse":
		result = float64(el.Data.MSpanInuse)
	case "MCacheSys":
		result = float64(el.Data.MCacheSys)
	case "Mallocs":
		result = float64(el.Data.Mallocs)
	case "NextGC":
		result = float64(el.Data.NextGC)
	case "OtherSys":
		result = float64(el.Data.OtherSys)
	case "PauseTotalNs":
		result = float64(el.Data.PauseTotalNs)
	case "StackInuse":
		result = float64(el.Data.StackInuse)
	case "StackSys":
		result = float64(el.Data.StackSys)
	case "Sys":
		result = float64(el.Data.Sys)
	case "TotalAlloc":
		result = float64(el.Data.TotalAlloc)
	case "NumForcedGC":
		result = float64(el.Data.NumForcedGC)
	case "NumGC":
		result = float64(el.Data.NumGC)
	case "GCCPUFraction":
		result = el.Data.GCCPUFraction
	default:
		return -1, fmt.Errorf("can not find metric name: %s", name)
	}

	//return Gauge(result), nil
	return result, nil
}
