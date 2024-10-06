package storage

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
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

func (s *MemStorage) SaveMetricsToDatabase(adr string) error {
	gauge := s.GetAllGauges()
	counter := s.GetAllCounters()

	db, err := sql.Open("postgres", adr)
	if err != nil {
		return fmt.Errorf("can't open database: %s", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("can't ping database: %s", err)
	}

	var maxID int64
	err = db.QueryRowContext(ctx, "SELECT COALESCE(MAX(load_id), 0) FROM metrics.metrics").Scan(&maxID)
	if err != nil {
		return fmt.Errorf("can't get max id: %s", err)
	}

	var query string
	for id, value := range gauge {
		query = "INSERT INTO metrics.metrics (load_id, metric_name, metric_type, gauge_value) VALUES ($1, $2, $3, $4)"
		_, err = db.ExecContext(ctx, query, maxID+1, id, "gauge", value)
		if err != nil {
			logger.Log.Error("can't insert gauge", zap.Error(err))
			return fmt.Errorf("can't insert gauge: %s", err)
		}
	}

	for id, value := range counter {
		query = "INSERT INTO metrics.metrics (load_id, metric_name, metric_type, counter_value) VALUES ($1, $2, $3, $4)"
		_, err = db.ExecContext(ctx, query, maxID+1, id, "counter", value)
		if err != nil {
			logger.Log.Error("can't insert counter", zap.Error(err))
			return fmt.Errorf("can't insert counter: %s", err)
		}
	}

	return nil
}
