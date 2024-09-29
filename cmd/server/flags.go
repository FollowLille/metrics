package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"strconv"
)

var (
	flagAddress       string
	flagLevel         string
	flagStoreInterval int64
	flagFilePath      string
	flagRestore       bool
)

func parseFlags() {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")
	pflag.Int64VarP(&flagStoreInterval, "store-interval", "i", 300, "store interval")
	pflag.StringVarP(&flagFilePath, "file-path", "f", "metrics.log", "file path")
	pflag.BoolVarP(&flagRestore, "restore", "r", true, "restore")

	pflag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		flagLevel = envLevel
	}

	envStoreInterval := os.Getenv("STORE_INTERVAL")
	if envStoreInterval != "" {
		storeInterval, err := strconv.ParseInt(envStoreInterval, 10, 64)
		if err != nil {
			fmt.Printf("invalid store interval: %s", envStoreInterval)
		}
		flagStoreInterval = storeInterval
	}

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		flagFilePath = envFilePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restore, err := strconv.ParseBool(envRestore)
		if err != nil {
			fmt.Printf("invalid restore value: %s", envRestore)
		}
		flagRestore = restore
	}

	fmt.Println("Flags:", flagAddress, flagLevel, flagStoreInterval, flagFilePath, flagRestore)
	fmt.Println("Address: ", os.Getenv("ADDRESS"))
	fmt.Println("Log level: ", os.Getenv("LOG_LEVEL"))
	fmt.Println("Store interval: ", os.Getenv("STORE_INTERVAL"))
	fmt.Println("File path: ", os.Getenv("FILE_STORAGE_PATH"))
	fmt.Println("Restore: ", os.Getenv("RESTORE"))
}
