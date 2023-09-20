package storage

import (
	"errors"
	"fmt"
	"strings"
)

type Gauge float64
type Counter int64

type Store struct {
	gauges      map[string]Gauge
	counters    map[string]Counter
	MetricsName []string
}

func NewStore() *Store {
	return &Store{
		gauges:   make(map[string]Gauge),
		counters: make(map[string]Counter),
		MetricsName: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
			"HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys",
			"Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"},
	}
}

func (m *Store) GetGauge(name string) (Gauge, bool) {
	val, exists := m.gauges[name]
	return val, exists
}

func (m *Store) SetGauge(name string, value Gauge) {
	m.gauges[name] = value
}

func (m *Store) GetCounter(name string) (Counter, bool) {
	val, exists := m.counters[name]
	return val, exists
}

func (m Store) UpdateCounter(name string, value Counter) {
	m.counters[name] += value
}

func (m Store) String(paramName string) (string, error) {
	var storList []string

	if paramName == "counters" {
		for k, v := range m.counters {
			storList = append(storList, k+":"+fmt.Sprintf("%d", v))
		}
	} else if paramName == "gauges" {
		for k, v := range m.gauges {
			storList = append(storList, k+":"+fmt.Sprintf("%f", v))
		}
	} else {
		return "", errors.New("bad param name")
	}

	return strings.Join(storList, ","), nil
}
