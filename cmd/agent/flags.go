package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

var flagAddress string
var flagReportInterval int64
var flagPollInterval int64

func parseFlags() error {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.Int64VarP(&flagReportInterval, "report-interval", "r", 10, "report interval")
	pflag.Int64VarP(&flagPollInterval, "poll-interval", "p", 2, "poll interval")
	pflag.Parse()
	if len(pflag.Args()) > 0 {
		return fmt.Errorf("unknown arguments: %v", flag.Args())
	}

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
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
	return nil
}
