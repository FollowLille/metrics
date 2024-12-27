// Package storage содержит хранилище метрик
package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"

	"github.com/FollowLille/metrics/internal/metrics"
)

// MemStorage хранилище метрик в памяти
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	muGauges sync.RWMutex
}

// NewMemStorage создает новый MemStorage
// и возвращает его в виде *MemStorage
//
// Возвращаемое значение:
//   - *MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		muGauges: sync.RWMutex{},
	}
}

// UpdateGauge обновляет значение метрики по имени
// Для работы с несколькими параллельными рутинами используется мьютекс
//
// Параметры:
//   - name - имя метрики
//   - value - значение метрики
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.muGauges.Lock()
	defer s.muGauges.Unlock()
	s.gauges[name] = value
}

// GetGauge возвращает значение метрики по имени
// Для работы с несколькими параллельными рутинами используется мьютекс
//
// Параметры:
//   - name - имя метрики
//
// Возвращаемое значение:
//   - float64 - значение метрики
//   - bool - существует ли метрика
func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.muGauges.RLock()
	defer s.muGauges.RUnlock()
	value, exists := s.gauges[name]
	return value, exists
}

// UpdateCounter обновляет значение счётчика по имени
//
// Параметры:
//   - name - имя счётчика
//   - value - значение счётчика
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counters[name] += value
}

// GetCounter возвращает значение счётчика по имени
//
// Параметры:
//   - name - имя счётчика
//
// Возвращаемое значение:
//   - int64 - значение счётчика
//   - bool - существует ли счётчик
func (s *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := s.counters[name]
	return value, exists
}

// Reset сбрасывает хранилище метрик
func (s *MemStorage) Reset() {
	s.gauges = make(map[string]float64)
	s.counters = make(map[string]int64)
}

// GetAllGauges возвращает все значения метрик
func (s *MemStorage) GetAllGauges() map[string]float64 {
	return s.gauges
}

// GetAllCounters возвращает все значения счётчиков
func (s *MemStorage) GetAllCounters() map[string]int64 {
	return s.counters
}

// GetAllMetrics возвращает все значения метрик
func (s *MemStorage) GetAllMetrics() map[string]interface{} {
	return map[string]interface{}{
		"gauges":   s.gauges,
		"counters": s.counters,
	}
}

// LoadMetricsFromFile загружает метрики из файла
//
// Параметры:
//   - file - файл с метриками
//
// Возвращаемое значение:
//   - error - ошибка загрузки
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
		return fmt.Errorf("can't read metrics from file: %w", err)
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

// SaveMetricsToFile сохраняет метрики в файл
//
// Параметры:
//   - file - файл для сохранения метрик
//
// Возвращаемое значение:
//   - error - ошибка сохранения
func (s *MemStorage) SaveMetricsToFile(file *os.File) error {
	metricsBatch := s.GetAllMetrics()
	data, err := json.Marshal(metricsBatch)
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
