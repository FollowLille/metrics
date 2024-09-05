package storage

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.gauges[name] = value
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := s.gauges[name]
	return value, exists
}
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counters[name] += value
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := s.counters[name]
	return value, exists
}

func (s *MemStorage) Reset() {
	s.gauges = make(map[string]float64)
	s.counters = make(map[string]int64)
}

func (s *MemStorage) GetAllGauges() map[string]float64 {
	return s.gauges
}

func (s *MemStorage) GetAllCounters() map[string]int64 {
	return s.counters
}
