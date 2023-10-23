package storage

import (
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/models"
	"github.com/rs/zerolog/log"
)

type Gauge float64
type Counter int64

type Store struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

func NewStore() *Store {
	return &Store{
		Gauges:   make(map[string]Gauge),
		Counters: make(map[string]Counter),
	}
}

func StoreMetricsToFile(s Storer) error {
	StoreFile, err := NewStoreFile(flags.FlagFileStorePath)
	if err != nil {
		return err
	}
	defer StoreFile.Close()

	if err := StoreFile.WriteMetrics(s.GetAllMetrics()); err != nil {
		return err
	}
	log.Info().Msg("metrics saved to file")

	return nil
}

func (m *Store) GetAllMetrics() (Store, error) {
	return *m, nil
}

func (m *Store) GetGauge(name string) (Gauge, bool, error) {
	val, exists := m.Gauges[name]
	return val, exists, nil
}

func (m *Store) GetGauges() (map[string]Gauge, error) {
	return m.Gauges, nil
}

func (m *Store) SetGauge(name string, value Gauge) error {
	m.Gauges[name] = value
	return nil
}

func (m *Store) GetCounter(name string) (Counter, bool, error) {
	val, exists := m.Counters[name]
	return val, exists, nil
}

func (m *Store) GetCounters() (map[string]Counter, error) {
	return m.Counters, nil
}

func (m *Store) UpdateCounter(name string, value Counter) error {
	m.Counters[name] += value
	return nil
}

func (m *Store) RestoreMetricsFromFile() error {
	RestoreFile, err := NewRestoreFile(flags.FlagFileStorePath)
	if err != nil {
		return err
	}
	defer RestoreFile.Close()

	if err := RestoreFile.ReadMetrics(m); err != nil {
		return err
	}

	return nil
}

func (m *Store) UpdateMetricBatch([]models.Metrics) error {
	return nil
}
