package main

import (
	"os"

	"github.com/spf13/pflag"
)

var (
	flagAddress string
	flagLevel   string
)

func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "port to listen on")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")

	pflag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		flagLevel = envLevel
	}
}
