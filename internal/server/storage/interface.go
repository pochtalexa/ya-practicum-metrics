package storage

type Storer interface {
	GetGauge(name string) (Gauge, bool)
	SetGauge(name string, value Gauge)
	GetCounter(name string) (Counter, bool)
	UpdateCounter(name string, value Counter)
}
