package storage

import (
	"github.com/pochtalexa/ya-practicum-metrics/internal/server/flags"
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

func (m *Store) GetGauge(name string) (Gauge, bool) {
	val, exists := m.Gauges[name]
	return val, exists
}

func (m *Store) GetGauges() map[string]Gauge {
	return m.Gauges
}

func (m *Store) SetGauge(name string, value Gauge) {
	m.Gauges[name] = value
}

func (m *Store) GetCounter(name string) (Counter, bool) {
	val, exists := m.Counters[name]
	return val, exists
}

func (m *Store) GetCounters() map[string]Counter {
	return m.Counters
}

func (m *Store) UpdateCounter(name string, value Counter) {
	m.Counters[name] += value
}

func (m *Store) StoreMetricsToFile() error {
	StoreFile, err := NewStoreFile(flags.FlagFileStorePath)
	if err != nil {
		return err
	}
	defer StoreFile.Close()

	if err := StoreFile.WriteMetrics(m); err != nil {
		return err
	}
	log.Info().Msg("metrics saved to file")

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
