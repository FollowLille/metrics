package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/FollowLille/metrics/internal/retry"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"

	"github.com/FollowLille/metrics/internal/config"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
)

type Agent struct {
	ServerAddress      string
	ServerPort         int64
	PollCount          int64
	PollInterval       time.Duration
	ReportSendInterval time.Duration
	metrics            map[string]float64
}

func NewAgent() *Agent {
	return &Agent{
		ServerAddress:      config.Address,
		ServerPort:         config.Port,
		PollInterval:       config.PollInterval,
		ReportSendInterval: config.ReportSendInterval,
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
	a.PollCount++
	m["PollCount"] = float64(a.PollCount)
	a.metrics = m
}

func (a *Agent) SendMetricsByBatch() error {
	logger.Log.Info("preparing to send metrics by batch")
	var m []metrics.Metrics
	for name, value := range a.metrics {
		var metric metrics.Metrics
		if name == "PollCount" {
			metric.MType = "counter"
			metric.ID = name
			delta := int64(value)
			metric.Delta = &delta
		} else {
			metric.MType = "gauge"
			metric.ID = name
			metric.Value = &value
		}
		m = append(m, metric)
	}
	logger.Log.Info("preparing to use metrics", zap.Any("metrics", m))
	fmt.Println("preparing to use metrics", m)

	jsonMetrics, err := json.Marshal(m)
	if err != nil {
		logger.Log.Error("failed to marshal metrics", zap.Error(err))
		return err
	}

	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		logger.Log.Error("failed to create gzip writer", zap.Error(err))
		return err
	}

	if _, err := gz.Write(jsonMetrics); err != nil {
		logger.Log.Error("failed to compress metric", zap.Error(err))
		return err
	}

	if err := gz.Flush(); err != nil {
		logger.Log.Error("failed to flush compressed data", zap.Error(err))
		return err
	}
	gz.Close()

	err = retry.Retry(func() error {
		return a.sendBatchMetrics(b)
	})

	if err != nil {
		logger.Log.Error("failed to send metrics", zap.Error(err))
		return err
	}
	return nil
}

func (a *Agent) sendBatchMetrics(b bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/updates", a.ServerAddress, a.ServerPort), &b)
	if err != nil {
		logger.Log.Error("failed to create request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

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

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("received non retriable status code", zap.Int("status_code", resp.StatusCode))
		return retry.ErrorNonRetriable
	}
	return nil
}

func (a *Agent) SendMetrics() error {
	for name, value := range a.metrics {
		var metric metrics.Metrics

		if name == "PollCount" {
			metric.MType = metrics.Counter
			metric.ID = name
			delta := int64(value)
			metric.Delta = &delta
		} else {
			metric.MType = metrics.Gauge
			metric.ID = name
			value := float64(value)
			metric.Value = &value
		}
		logger.Log.Info("preparing to use metric", zap.Any("metric", metric))

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
			return a.sendSingleMetric(b)
		})
		if err != nil {
			logger.Log.Error("failed to send metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
			return err
		}
	}
	return nil
}

func (a *Agent) sendSingleMetric(b bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%d/update", a.ServerAddress, a.ServerPort), &b)
	if err != nil {
		logger.Log.Error("failed to create request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

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

func (a *Agent) Run() {
	logger.Log.Info("agent running")
	pollTicker := time.NewTicker(a.PollInterval)
	reportTicker := time.NewTicker(a.ReportSendInterval)

	for {
		select {
		case <-pollTicker.C:
			a.GetMetrics()
		case <-reportTicker.C:
			err := a.SendMetricsByBatch()
			if err != nil {
				logger.Log.Error("failed to send metrics", zap.Error(err))
			}
		}
	}
}
