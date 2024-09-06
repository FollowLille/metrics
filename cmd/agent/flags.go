package main

import (
	"flag"
	"fmt"
)

var flagPort int64
var flagReportInterval int64
var flagPollInterval int64

func parseFlags() error {
	flag.Int64Var(&flagPort, "a", 8080, "port to listen on")
	flag.Int64Var(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()
	if len(flag.Args()) > 0 {
		return fmt.Errorf("неизвестные аргументы: %v", flag.Args())
	}
	return nil
}
