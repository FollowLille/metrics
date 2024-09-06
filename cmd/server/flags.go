package main

import (
	"github.com/spf13/pflag"
)

var flagPort int64

func parseFlags() {
	pflag.Int64VarP(&flagPort, "port", "a", 8080, "port to listen on")
	pflag.Parse()
}
