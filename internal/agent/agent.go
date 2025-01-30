// Package agent содержит логику работы агента
// Агент слушает очередь с метриками и отправляет их на сервер
package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/config"
	"github.com/FollowLille/metrics/internal/crypto"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
	"github.com/FollowLille/metrics/internal/retry"
)

type Agent struct {
	ServerAddress      string
	HashKey            string
	ServerPort         int64
	PollCount          int64
	RateLimit          int64
	PollInterval       time.Duration
	ReportSendInterval time.Duration
	PublicKey          *rsa.PublicKey
	metrics            map[string]float64
	mutex              sync.Mutex
}

// NewAgent инициализирует агента
func NewAgent() *Agent {
	return &Agent{
		ServerAddress:      config.Address,
		ServerPort:         config.Port,
		PollInterval:       config.PollInterval,
		ReportSendInterval: config.ReportSendInterval,
		RateLimit:          config.RateLimit,
		metrics:            make(map[string]float64),
	}
}

// ChangeIntervalByName изменяет интервал по имени
//
// Параметры:
//   - name - имя интервала
//   - seconds - интервал в секундах
//
// Возвращаемое значение:
//   - error - в случае ошибки
func (a *Agent) ChangeIntervalByName(name string, seconds int64) error {
	if seconds < 1 {
		return fmt.Errorf("incorect inverval value: %d, value must be > 0", seconds)
	}
	newInterval := time.Second * time.Duration(seconds)
	if name == "poll" {
		a.PollInterval = newInterval
	} else if name == "report" {
		a.ReportSendInterval = newInterval
	} else {
		return fmt.Errorf("invalid interval name: %s", name)
	}
	return nil
}

// ChangeAddress изменяет адрес
//
// Параметры:
//   - address - адрес
//
// Возвращаемое значение:
//   - error
func (a *Agent) ChangeAddress(address string) error {
	u, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}
	if u.Port() != "" {
		return fmt.Errorf("invalid address: %s, port must be empty", address)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("invalid address: %s", address)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("некорректный адрес: %s", address)
	}
	a.ServerAddress = u.Hostname()
	return nil
}

// ChangePort изменяет порт
//
// Параметры:
//   - port - порт
//
// Возвращаемое значение:
//   - error
func (a *Agent) ChangePort(port int64) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	a.ServerPort = port
	return nil
}

// GetMetrics получает метрики
func (a *Agent) GetMetrics() {
	m := metrics.GetRuntimeMetrics()
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for k, v := range m {
		a.metrics[k] = v
	}
}

// GetGopsutilMetrics получает расширенные метрики
func (a *Agent) GetGopsutilMetrics() {
	m := metrics.GetGopsutilMetrics()
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for k, v := range m {
		a.metrics[k] = v
	}
}

// IncreasePollCount увеличивает счетчик опросов
func (a *Agent) IncreasePollCount() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.PollCount++
}

// Run запускает агента
func (a *Agent) Run() {
	logger.Log.Info("agent running")
	logger.Log.Info("Intervals: ", zap.String("poll", a.PollInterval.String()), zap.String("report", a.ReportSendInterval.String()))
	pollTicker := time.NewTicker(a.PollInterval)
	reportTicker := time.NewTicker(a.ReportSendInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			go a.GetMetrics()
			go a.GetGopsutilMetrics()
			go a.IncreasePollCount()
		case <-reportTicker.C:
			go a.ParallelSendMetrics()
		}
	}
}

// ParallelSendMetrics запускает параллельное отправление метрик
func (a *Agent) ParallelSendMetrics() {
	metricsChan := make(chan metrics.Metrics, 30)

	for i := int64(0); i < a.RateLimit; i++ {
		go a.sendByWorker(metricsChan)
	}
	a.mutex.Lock()

	var m metrics.Metrics
	for name, value := range a.metrics {
		m.MType = metrics.Gauge
		m.ID = name
		m.Value = &value
		metricsChan <- m
	}
	m.MType = metrics.Counter
	m.ID = "PollCount"
	delta := a.PollCount
	m.Delta = &delta
	metricsChan <- m

	a.mutex.Unlock()

	close(metricsChan)
}

// sendByWorker отправляет метрики по каналу
// Принимает канал метрик
//
// Параметры:
//   - metricsChan - канал метрик
func (a *Agent) sendByWorker(metricsChan <-chan metrics.Metrics) {
	for m := range metricsChan {
		err := a.sendSingleMetric(m)
		if err != nil {
			logger.Log.Error("failed to send single metric", zap.Error(err))
		}
	}
}

// sendSingleMetric отправляет метрику
// Принимает метрику
//
// Параметры:
//   - metric - метрика
//
// Возвращаемое значение:
//   - error
func (a *Agent) sendSingleMetric(metric metrics.Metrics) error {
	jsonMetrics, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error("failed to marshal metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}

	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		logger.Log.Error("failed to create gzip writer", zap.Error(err))
		return err
	}

	if _, err := gz.Write(jsonMetrics); err != nil {
		logger.Log.Error("failed to compress metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}

	if err := gz.Flush(); err != nil {
		logger.Log.Error("failed to flush compressed data", zap.Error(err))
		return err
	}
	gz.Close()
	err = retry.Retry(func() error {
		return a.sendRequest(b)
	})
	if err != nil {
		logger.Log.Error("failed to send metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
		return err
	}
	return nil
}

// sendRequest отправляет запрос на сервер
// если включен шифрование, то шифруем данные
// если включен хеш, то вычисляем хеш и добавляем его в заголовок
//
// Параметры:
//   - b bytes.Buffer - буфер с данными
//
// Возвращаемое значение:
//   - error
func (a *Agent) sendRequest(b bytes.Buffer) error {
	data := b.Bytes()

	if a.PublicKey != nil {
		encryptedData, err := crypto.Encrypt(a.PublicKey, data)
		if err != nil {
			logger.Log.Error("failed to encrypt data", zap.Error(err))
			return err
		}
		data = encryptedData
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/update", a.ServerAddress, a.ServerPort), bytes.NewReader(data))
	if err != nil {
		logger.Log.Error("failed to create request", zap.Error(err))
		return err
	}
	if a.PublicKey != nil {
		req.Header.Set("Content-Type", "application/octet-stream")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Content-Encoding", "gzip")

	if a.HashKey != "" {
		hash := crypto.CalculateHash([]byte(a.HashKey), b.Bytes())
		req.Header.Set("HashSHA256", hash)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error("failed to send request", zap.Error(err))
		return retry.ErrorConnection
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 && resp.StatusCode <= 504 {
		logger.Log.Error("received retriable status code", zap.Int("status_code", resp.StatusCode))
		return retry.ErrorServer
	}

	return nil
}
