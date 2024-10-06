package config

import (
	"net/http"
	"time"
)

const (
	PollInterval       = 2 * time.Second
	ReportSendInterval = 10 * time.Second
	Address            = "localhost"
	Port               = 8080
	ContentType        = "text/plain"
)

// status constants
const (
	StatusOk         = http.StatusOK
	StatusBadRequest = http.StatusBadRequest
	StatusNotFound   = http.StatusNotFound
	ServerError      = http.StatusInternalServerError
)
