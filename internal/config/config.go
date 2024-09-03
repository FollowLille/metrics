package config

import (
	"time"
)

const (
	PollInterval       = 2 * time.Second
	ReportSendInterval = 10 * time.Second
	Address            = "http://localhost"
	Port               = 8080
	ContentType        = "text/plain"
)
