package main

import (
	"os"

	"github.com/spf13/pflag"
)

var flagAddress string
var flagDatabaseAddress string

func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address and port to listen on")
	pflag.StringVarP(&flagDatabaseAddress, "db", "d", "localhost:5432", "database address and port to connect to")
	pflag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}

	if envDB := os.Getenv("DATABASE_DSN"); envDB != "" {
		flagDatabaseAddress = envDB
	}
}
