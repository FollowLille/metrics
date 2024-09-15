package main

import (
	"os"

	"github.com/spf13/pflag"
)

var flagAddress string

func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "port to listen on")
	pflag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}
}
