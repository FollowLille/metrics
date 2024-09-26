package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

var (
	flagAddress string
	flagLevel   string
)

func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")

	pflag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		flagLevel = envLevel
	}
	fmt.Println("Flags:", flagAddress, flagLevel)
	fmt.Println("Address: ", os.Getenv("ADDRESS"))
}
