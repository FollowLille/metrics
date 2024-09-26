package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
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

func (a *Agent) SendMetrics() error {
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

		logger.Log.Info("preparing to use metric", zap.Any("metric", metric))

		jsonMetrics, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Error("failed to marshal metric", zap.String("metric", fmt.Sprintf("%+v", metric)), zap.Error(err))
			return err
		}

		addr := fmt.Sprintf("http://%s:%d/update", a.ServerAddress, a.ServerPort)

		logger.Log.Info("sending metric", zap.String("url", addr), zap.ByteString("jsonMetrics", jsonMetrics))

		resp, err := http.Post(addr, "application/json", bytes.NewReader(jsonMetrics))
		if err != nil {
			logger.Log.Error("failed to send metrics", zap.String("url", addr), zap.Error(err))
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != config.StatusOk {
			body, _ := io.ReadAll(resp.Body)
			logger.Log.Error("invalid status code", zap.Int("status_code", resp.StatusCode), zap.String("body", string(body)))
			return fmt.Errorf("invalid status code: %d", resp.StatusCode)
		}
	}
	return nil
}

func (a *Agent) Run() {
	fmt.Printf("agent started to work with http://%s:%d/update", a.ServerAddress, a.ServerPort)
	pollTicker := time.NewTicker(a.PollInterval)
	reportTicker := time.NewTicker(a.ReportSendInterval)

	for {
		select {
		case <-pollTicker.C:
			a.GetMetrics()
			fmt.Println(a.metrics)
		case <-reportTicker.C:
			err := a.SendMetrics()
			if err != nil {
				panic(err)
			}
		}
	}
}
