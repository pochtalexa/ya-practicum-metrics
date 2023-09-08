package storage


type Gauge float64
type Counter float64

type MemStore struct {    
    gauges   map[string]Gauge
	counters map[string]Counter
}

type MemStorer interface {
	GetGauge(name string)(Gauge, bool)
	SetGauge(name string, value Gauge)
    GetCounter(name string)(Counter, bool)
	UpdateCounter(name string, value Counter)	
}

func NewMemStore() *MemStore {
    return &MemStore{gauges: make(map[string]Gauge), counters: make(map[string]Counter)}
}

func (m MemStore) GetGauge(name string) (Gauge, bool) {
    val, exists := m.gauges[name]
    return val, exists
}

func (m MemStore) SetGauge(name string, value Gauge) {
    m.gauges[name] = value
}

func (m MemStore) GetCounter(name string) (Counter, bool) {
    val, exists := m.counters[name] 
    return val, exists
}

func (m MemStore) UpdateCounter(name string, value Counter) {
    m.counters[name] = value
}
