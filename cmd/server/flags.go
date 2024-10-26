package main

import (
	"go.uber.org/zap"
	"os"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/FollowLille/metrics/internal/logger"
)

var (
	flagStoreInterval   int64
	flagAddress         string
	flagLevel           string
	flagFilePath        string
	flagRestoreStr      string
	flagDatabaseAddress string
	flagStorePlace      string
	flagHashKey         string
	flagRestore         bool
)

func parseFlags() {
	pflag.Int64VarP(&flagStoreInterval, "store-interval", "i", 300, "store interval")
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")
	pflag.StringVarP(&flagFilePath, "file-path", "f", "", "file path")
	pflag.StringVarP(&flagRestoreStr, "restore", "r", "true", "restore")
	pflag.StringVarP(&flagDatabaseAddress, "database-address", "d", "", "database address")
	pflag.StringVarP(&flagHashKey, "hash-key", "k", "secret", "hash key")

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
			logger.Log.Error("Invalid store interval value", zap.Error(err))
			os.Exit(1)
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

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		flagHashKey = envHashKey
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
		logger.Log.Error("Invalid restore value", zap.Error(err))
		os.Exit(1)
	}

	logger.Log.Info("Flags", zap.Int64("store-interval", flagStoreInterval), zap.String("hash-key", flagHashKey), zap.String("address", flagAddress), zap.String("level", flagLevel), zap.String("file-path", flagFilePath), zap.String("restore", flagRestoreStr))

}
