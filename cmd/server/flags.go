package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"strconv"
)

var (
	flagStoreInterval   int64
	flagAddress         string
	flagLevel           string
	flagFilePath        string
	flagRestoreStr      string
	flagDatabaseAddress string
	flagRestore         bool
)

func parseFlags() {
	pflag.Int64VarP(&flagStoreInterval, "store-interval", "i", 300, "store interval")
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")
	pflag.StringVarP(&flagFilePath, "file-path", "f", "./metrics", "file path")
	pflag.StringVarP(&flagRestoreStr, "restore", "r", "true", "restore")
	pflag.StringVarP(&flagDatabaseAddress, "database-address", "d", "postgres://login:password@localhost:5432/for_go?sslmode=disable", "database address")

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
		flagRestoreStr = envRestore
	}

	if envDatabaseAddress := os.Getenv("DATABASE_DSN"); envDatabaseAddress != "" {
		flagDatabaseAddress = envDatabaseAddress
	}

	var err error
	flagRestore, err = strconv.ParseBool(flagRestoreStr)
	if err != nil {
		fmt.Printf("invalid restore value: %s", flagRestoreStr)
		os.Exit(1)
	}

	fmt.Println("Flags:", flagAddress, flagLevel, flagStoreInterval, flagFilePath, flagRestore, flagDatabaseAddress)
	fmt.Println("Address: ", os.Getenv("ADDRESS"))
	fmt.Println("Log level: ", os.Getenv("LOG_LEVEL"))
	fmt.Println("Store interval: ", os.Getenv("STORE_INTERVAL"))
	fmt.Println("File path: ", os.Getenv("FILE_STORAGE_PATH"))
	fmt.Println("Restore: ", os.Getenv("RESTORE"))
	fmt.Println("Database address: ", os.Getenv("DATABASE_DSN"))
}
