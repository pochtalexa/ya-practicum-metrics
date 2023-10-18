package storage

import "github.com/pochtalexa/ya-practicum-metrics/internal/server/models"

type Storer interface {
	GetGauge(name string) (Gauge, bool)
	GetGauges() map[string]Gauge
	SetGauge(name string, value Gauge)
	GetCounter(name string) (Counter, bool)
	GetCounters() map[string]Counter
	UpdateCounter(name string, value Counter)
	GetAllMetrics() Store
	UpdateMetricBatch([]models.Metrics) error
}
