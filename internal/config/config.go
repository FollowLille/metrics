package config

import (
	"time"
)

const (
	PollInterval       = 2 * time.Second
	ReportSendInterval = 10 * time.Second
	Address            = "localhost"
	Port               = 8080
	RateLimit          = 10
)

var DatabaseRetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}
