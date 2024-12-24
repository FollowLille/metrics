package agent

import (
	"bytes"
	"compress/gzip"
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
	metrics            map[string]float64
	mutex              sync.Mutex
}

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

func (a *Agent) ChangePort(port int64) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	a.ServerPort = port
	return nil
}

func (a *Agent) GetMetrics() {
	m := metrics.GetRuntimeMetrics()
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for k, v := range m {
		a.metrics[k] = v
	}
}

func (a *Agent) GetGopsutilMetrics() {
	m := metrics.GetGopsutilMetrics()
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for k, v := range m {
		a.metrics[k] = v
	}
}

func (a *Agent) IncreasePollCount() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.PollCount++
	a.metrics["PollCount"] = float64(a.PollCount)
}

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
			logger.Log.Info("sent metrics", zap.Int64("count", a.PollCount))
			go a.ParallelSendMetrics()
		}
	}
}

func (a *Agent) ParallelSendMetrics() {
	metricsChan := make(chan metrics.Metrics)

	for i := int64(0); i < a.RateLimit; i++ {
		go a.sendByWorker(metricsChan)
	}
	a.mutex.Lock()
	for name, value := range a.metrics {
		var m metrics.Metrics
		if name == "PollCount" {
			m.MType = metrics.Counter
			m.ID = name
			delta := int64(value)
			m.Delta = &delta
		} else {
			m.MType = metrics.Gauge
			m.ID = name
			m.Value = &value
		}
		metricsChan <- m
	}
	a.mutex.Unlock()

	close(metricsChan)
}

func (a *Agent) sendByWorker(metricsChan <-chan metrics.Metrics) {
	for m := range metricsChan {
		err := a.sendSingleMetric(m)
		if err != nil {
			logger.Log.Error("failed to send single metric", zap.Error(err))
		}
	}
}

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

func (a *Agent) sendRequest(b bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/update", a.ServerAddress, a.ServerPort), &b)
	if err != nil {
		logger.Log.Error("failed to create request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

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
