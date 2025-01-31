// Package main отвечает за инициализацию и запуск сервера
// Он включает в себя функции для парсинга командных флагов и переменных окружения,
// а также настройку логгирования.
package main

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/FollowLille/metrics/internal/logger"
)

// Структура файла с флагами для инициализации через json
type Config struct {
	StoreInterval   int64  `json:"store_interval"`
	Address         string `json:"address"`
	Level           string `json:"level"`
	FilePath        string `json:"file_path"`
	RestoreStr      string `json:"restore_str"`
	DatabaseAddress string `json:"database_address"`
	StorePlace      string `json:"store_place"`
	HashKey         string `json:"hash_key"`
	CryptoKeyPath   string `json:"crypto_key"`
	Restore         string `json:"restore"`
}

// Флаги
var (
	flagStoreInterval   int64  // интервал хранения данных
	flagAddress         string // адрес для прослушивания
	flagLevel           string // уровень логирования
	flagFilePath        string // путь к файлу логирования
	flagRestoreStr      string // флаг восстановления
	flagDatabaseAddress string // адрес базы данных
	flagStorePlace      string // место хранения
	flagHashKey         string // ключ хэша
	flagCryptoKeyPath   string // путь к файлу с приватным ключом
	flagConfigFilePath  string // путь к файлу с конфигом
	flagRestore         bool   // флаг восстановления
)

// parseFlags парсит командные флаги и переменные окружения для настройки сервера.
// Флаги включают адрес сервера, адрес базы данных, адрес системы начисления и уровень логирования.
// Если переменные окружения определены, они имеют приоритет над значениями по умолчанию.
func parseFlags() {
	pflag.Int64VarP(&flagStoreInterval, "store-interval", "i", 300, "store interval")
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log level")
	pflag.StringVarP(&flagFilePath, "file-path", "f", "", "file path")
	pflag.StringVarP(&flagRestoreStr, "restore", "r", "true", "restore")
	pflag.StringVarP(&flagDatabaseAddress, "database-address", "d", "", "database address")
	pflag.StringVarP(&flagCryptoKeyPath, "crypto-key", "y", "", "private key path")
	pflag.StringVarP(&flagConfigFilePath, "config", "c", "", "path to config file")
	pflag.StringVarP(&flagHashKey, "hash-key", "k", "", "hash key")

	pflag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		flagLevel = envLevel
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		flagCryptoKeyPath = envCryptoKey
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

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		flagConfigFilePath = envConfig
	}

	if flagConfigFilePath != "" {
		err := loadConfigFromFile(flagConfigFilePath)
		if err != nil {
			logger.Log.Error("Invalid config file", zap.Error(err))
			os.Exit(1)
		}
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

	logger.Log.Info("Flags",
		zap.Int64("store-interval", flagStoreInterval),
		zap.String("hash-key", flagHashKey),
		zap.String("address", flagAddress),
		zap.String("level", flagLevel),
		zap.String("file-path", flagFilePath),
		zap.String("restore", flagRestoreStr),
		zap.String("database-address", flagDatabaseAddress),
		zap.String("crypto-key", flagCryptoKeyPath),
		zap.String("store-place", flagStorePlace),
	)
}

func loadConfigFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var cfg Config

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	if cfg.StoreInterval != 0 {
		flagStoreInterval = cfg.StoreInterval
	}
	if cfg.Address != "" {
		flagAddress = cfg.Address
	}
	if cfg.Level != "" {
		flagLevel = cfg.Level
	}
	if cfg.FilePath != "" {
		flagFilePath = cfg.FilePath
	}
	flagRestore, err = strconv.ParseBool(flagRestoreStr)
	if err != nil {
		logger.Log.Error("Invalid restore value", zap.Error(err))
		os.Exit(1)
	}
	if cfg.DatabaseAddress != "" {
		flagDatabaseAddress = cfg.DatabaseAddress
	}
	if cfg.CryptoKeyPath != "" {
		flagCryptoKeyPath = cfg.CryptoKeyPath
	}
	return nil
}
