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
	flagStorePlace      string
	flagRestore         bool
)

func parseFlags() {
	pflag.Int64VarP(&flagStoreInterval, "store-interval", "i", 300, "store interval")
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")
	pflag.StringVarP(&flagFilePath, "file-path", "f", "", "file path")
	pflag.StringVarP(&flagRestoreStr, "restore", "r", "true", "restore")
	pflag.StringVarP(&flagDatabaseAddress, "database-address", "d", "", "database address")

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

	if flagDatabaseAddress != "" {
		flagStorePlace = "database"
	} else if flagFilePath != "" {
		flagStorePlace = "file"
	} else {
		flagStorePlace = "memory"
	}

	var err error
	flagRestore, err = strconv.ParseBool(flagRestoreStr)
	if err != nil {
		fmt.Printf("invalid restore value: %s", flagRestoreStr)
		os.Exit(1)
	}

	fmt.Println("Flags:", flagAddress, flagLevel, flagStoreInterval, flagFilePath, flagRestore, flagDatabaseAddress)
	fmt.Println("Address: ", flagAddress)
	fmt.Println("Log level: ", flagLevel)
	fmt.Println("Store interval: ", flagStoreInterval)
	fmt.Println("File path: ", flagFilePath)
	fmt.Println("Restore: ", flagRestore)
	fmt.Println("Database address: ", flagDatabaseAddress)
}
