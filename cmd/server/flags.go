package main

import (
	"github.com/spf13/pflag"
)

var flagAddress string

func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "port to listen on")
	pflag.Parse()
}
