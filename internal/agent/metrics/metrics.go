package metrics

import (
	"math"
	"math/rand"
)

type gauge float64
type counter int64

type RuntimeMetrics struct {
	MetricsName []string
	Counters    map[string]counter
	PollCount   counter
	RandomValue gauge
}

func New() *RuntimeMetrics {
	return &RuntimeMetrics{
		MetricsName: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
			"HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys",
			"Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"},

		Counters: make(map[string]counter),
	}
}

func (el *RuntimeMetrics) UpdateCounter(name string, value counter) {
	el.Counters[name] = value
}

func (el *RuntimeMetrics) PollCountInc() {
	el.PollCount++
}

func (el *RuntimeMetrics) PollCountDrop() {
	el.PollCount = 0
}

func (el *RuntimeMetrics) RandomValueUpdate() {
	el.RandomValue = gauge(rand.Float64() * math.Pow(10, 6))
}
