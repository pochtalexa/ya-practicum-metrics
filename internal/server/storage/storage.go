package storage

type Gauge float64
type Counter int64

type Store struct {
	Gauges      map[string]Gauge
	Counters    map[string]Counter
	MetricsName []string
}

func NewStore() *Store {
	return &Store{
		Gauges:   make(map[string]Gauge),
		Counters: make(map[string]Counter),
		MetricsName: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
			"HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys",
			"Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"},
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
