package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/FollowLille/metrics/internal/metrics"
	"os"
)

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

func (s *MemStorage) GetAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"gauges":   s.gauges,
		"counters": s.counters,
	}
}

func (s *MemStorage) LoadMetricsFromFile(file *os.File) error {
	scanner := bufio.NewScanner(file)
	var lastValue string

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lastValue = line
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("can't read metrics from file: %s", err)
	}

	if lastValue == "" {
		return nil
	}

	var metricsFile metrics.MetricsFile

	err := json.Unmarshal([]byte(lastValue), &metricsFile)
	if err != nil {
		return fmt.Errorf("can't unmarshal metrics: %s", err)
	}

	for id, value := range metricsFile.Gauges {
		s.UpdateGauge(id, value)
	}

	for id, delta := range metricsFile.Counters {
		s.UpdateCounter(id, delta)
	}

	return nil
}

func (s *MemStorage) SaveMetricsToFile(file *os.File) error {
	metrics := s.GetAllMetrics()
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("can't marshal metrics: %s", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("can't write metrics to file: %s", err)
	}
	_, err = file.WriteString("\n")
	if err != nil {
		return fmt.Errorf("can't write metrics to file: %s", err)
	}
	return nil
}
