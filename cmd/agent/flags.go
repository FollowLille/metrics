package main

import (
	"flag"
	"fmt"
	"github.com/spf13/pflag"
)

var flagAddress string
var flagReportInterval int64
var flagPollInterval int64

func parseFlags() error {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "port to listen on")
	pflag.Int64VarP(&flagReportInterval, "report-interval", "r", 10, "report interval")
	pflag.Int64VarP(&flagPollInterval, "poll-interval", "p", 2, "poll interval")
	pflag.Parse()
	if len(pflag.Args()) > 0 {
		return fmt.Errorf("неизвестные аргументы: %v", flag.Args())
	}
	return nil
}
