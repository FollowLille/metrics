// Package metrics содержит функции для получения метрик
package metrics

import (
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const Counter = "counter" // счетчик
const Gauge = "gauge"     // метрики

// GetRuntimeMetrics возвращает метрики процессора и памяти
//
// Возвращаемое значение:
//   - метрики процессора и памяти
func GetRuntimeMetrics() map[string]float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	metrics := make(map[string]float64)
	metrics["Alloc"] = float64(m.Alloc)
	metrics["BuckHashSys"] = float64(m.BuckHashSys)
	metrics["Frees"] = float64(m.Frees)
	metrics["GCCPUFraction"] = m.GCCPUFraction
	metrics["GCSys"] = float64(m.GCSys)
	metrics["HeapAlloc"] = float64(m.HeapAlloc)
	metrics["HeapIdle"] = float64(m.HeapIdle)
	metrics["HeapInuse"] = float64(m.HeapInuse)
	metrics["HeapObjects"] = float64(m.HeapObjects)
	metrics["HeapReleased"] = float64(m.HeapReleased)
	metrics["HeapSys"] = float64(m.HeapSys)
	metrics["LastGC"] = float64(m.LastGC)
	metrics["Lookups"] = float64(m.Lookups)
	metrics["MCacheInuse"] = float64(m.MCacheInuse)
	metrics["MCacheSys"] = float64(m.MCacheSys)
	metrics["MSpanInuse"] = float64(m.MSpanInuse)
	metrics["MSpanSys"] = float64(m.MSpanSys)
	metrics["Mallocs"] = float64(m.Mallocs)
	metrics["NextGC"] = float64(m.NextGC)
	metrics["NumForcedGC"] = float64(m.NumForcedGC)
	metrics["NumGC"] = float64(m.NumGC)
	metrics["OtherSys"] = float64(m.OtherSys)
	metrics["PauseTotalNs"] = float64(m.PauseTotalNs)
	metrics["StackInuse"] = float64(m.StackInuse)
	metrics["StackSys"] = float64(m.StackSys)
	metrics["Sys"] = float64(m.Sys)
	metrics["TotalAlloc"] = float64(m.TotalAlloc)
	metrics["RandomValue"] = rand.Float64()

	return metrics
}

// GetGopsutilMetrics возвращает метрики процессора и памяти через gopsutil
//
// Возвращаемое значение:
//   - метрики процессора и памяти
func GetGopsutilMetrics() map[string]float64 {
	metrics := make(map[string]float64)

	// Получаем данные о памяти
	v, err := mem.VirtualMemory()
	if err == nil {
		metrics["TotalMemory"] = float64(v.Total)
		metrics["FreeMemory"] = float64(v.Free)
	}

	// Получаем данные о загрузке CPU
	cpuUtilization, err := cpu.Percent(0, false)
	if err == nil && len(cpuUtilization) > 0 {
		metrics["CPUutilization1"] = cpuUtilization[0]
	}

	return metrics
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type MetricsFile struct {
	Counters map[string]int64   `json:"counters"`
	Gauges   map[string]float64 `json:"gauges"`
}
