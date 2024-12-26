// Package config содержит конфигурацию приложения
package config

import (
	"time"
)

const (
	PollInterval       = 2 * time.Second  // интервал опроса метрик
	ReportSendInterval = 10 * time.Second // интервал отправки метрик
	Address            = "localhost"      // адрес для прослушивания
	Port               = 8080             // порт для прослушивания
	RateLimit          = 3                // лимит на кол-во одновременных воркеров
)

// DatabaseRetryDelays - задержки между повторными попытками подключения к базе данных
var DatabaseRetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}
