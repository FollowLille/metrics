package main

import "flag"

var flagPort int64

func parseFlags() {
	flag.Int64Var(&flagPort, "a", 8080, "port to listen on")
	flag.Parse()
}
