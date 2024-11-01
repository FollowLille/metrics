package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/logger"
)

var flagAddress string
var flagHashKey string
var flagReportInterval int64
var flagPollInterval int64
var flagRateLimit int64

func parseFlags() error {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagHashKey, "hash-key", "k", "", "hash key")
	pflag.Int64VarP(&flagReportInterval, "report-interval", "r", 10, "report interval")
	pflag.Int64VarP(&flagPollInterval, "poll-interval", "p", 2, "poll interval")
	pflag.Int64VarP(&flagRateLimit, "rate-limit", "l", 4, "rate limit")
	pflag.Parse()
	if len(pflag.Args()) > 0 {
		return fmt.Errorf("unknown arguments: %v", flag.Args())
	}

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		flagHashKey = envHashKey
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		interval, err := strconv.ParseInt(envReportInterval, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid interval value: %d", interval)
		}
		flagReportInterval = interval
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		interval, err := strconv.ParseInt(envPollInterval, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid interval value: %d", interval)
		}
		flagPollInterval = interval
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		interval, err := strconv.ParseInt(envRateLimit, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid interval value: %d", interval)
		}
		flagRateLimit = interval
	}

	logger.Log.Info("Flags", zap.String("address", flagAddress), zap.String("hash-key", flagHashKey), zap.Int64("report-interval", flagReportInterval), zap.Int64("poll-interval", flagPollInterval))
	return nil

}
