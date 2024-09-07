package agent

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/FollowLille/metrics/internal/config"
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
		return fmt.Errorf("incorrrect address: %s, port must be empty", address)
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
		metricType := "gauge"
		if name == "PollCount" {
			metricType = "counter"

			addr := fmt.Sprintf("http://%s:%d/update/%s/%s/%f", a.ServerAddress, a.ServerPort, metricType, name, value)
			resp, err := http.Post(addr, config.ContentType, nil)
			if err != nil {
				return err
			}
			if err := resp.Body.Close(); err != nil {
				return err
			}

			if resp.StatusCode != 200 {
				return fmt.Errorf("invalid status code: %d", resp.StatusCode)
			}
		}
	}
	return nil
}

func (a *Agent) Run() {
	fmt.Println("agent started")
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
