// Package main для запуска агента
// Данная часть содержит информацию об использемых флагах
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/logger"
)

// Структура файла с флагами для инициализации через json
type Config struct {
	Address        string `json:"address"`
	HashKey        string `json:"hash_key"`
	CryptoKeyPath  string `json:"crypto_key"`
	ReportInterval int64  `json:"report_interval"`
	PollInterval   int64  `json:"poll_interval"`
	RateLimit      int64  `json:"rate_limit"`
}

// Флаги
var (
	flagAddress        string // адрес для прослушивания
	flagHashKey        string // ключ хэша
	flagCryptoKeyPath  string // путь к файлу с ключом
	flagConfigFilePath string // путь к файлу с конфигом
	flagPollInterval   int64  // интервал опроса
	flagReportInterval int64  // интервал отчета
	flagRateLimit      int64  // лимит на кол-во одновременных воркеров
)

// parseFlags парсит командные флаги и переменные окружения для настройки сервера.
// Флаги включают адрес сервера, адрес базы данных, адрес системы начисления и уровень логирования.
// Если переменные окружения определены, они имеют приоритет над значениями по умолчанию.
//
// Пример использования:
//
//			-address=127.0.0.1:8080
//	     	-hash-key=secret
//			-crypto-key=/path/to/file
//			-сonfig=cfg.json
//			-report-interval=10
//			-poll-interval=2
//			-rate-limit=4
//
// После парсинга флагов, информация о них логируется с использованием zap.
func parseFlags() error {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagHashKey, "hash-key", "k", "", "hash key")
	pflag.StringVarP(&flagCryptoKeyPath, "crypto-key", "y", "", "path to crypto key file")
	pflag.StringVarP(&flagConfigFilePath, "config", "c", "", "path to config file")
	pflag.Int64VarP(&flagReportInterval, "report-interval", "r", 10, "report interval")
	pflag.Int64VarP(&flagPollInterval, "poll-interval", "p", 2, "poll interval")
	pflag.Int64VarP(&flagRateLimit, "rate-limit", "l", 4, "rate limit")
	pflag.Parse()
	if len(pflag.Args()) > 0 {
		return fmt.Errorf("unknown arguments: %v", flag.Args())
	}

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		flagAddress = envAddress
	}

	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		flagHashKey = envHashKey
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		flagCryptoKeyPath = envCryptoKey
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		flagConfigFilePath = envConfig
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		interval, err := strconv.ParseInt(envReportInterval, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid interval value: %d", interval)
		}
		flagReportInterval = interval
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		interval, err := strconv.ParseInt(envPollInterval, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid interval value: %d", interval)
		}
		flagPollInterval = interval
	}

	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		interval, err := strconv.ParseInt(envRateLimit, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid interval value: %d", interval)
		}
		flagRateLimit = interval
	}

	if flagConfigFilePath != "" {
		err := loadConfigFromFile(flagConfigFilePath)
		if err != nil {
			return err
		}
	}

	logger.Log.Info("Flags",
		zap.String("address", flagAddress),
		zap.String("hash-key", flagHashKey),
		zap.String("crypto-key", flagCryptoKeyPath),
		zap.String("config", flagConfigFilePath),
		zap.Int64("report-interval", flagReportInterval),
		zap.Int64("poll-interval", flagPollInterval),
		zap.Int64("rate-limit", flagRateLimit),
	)
	return nil
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
	if cfg.Address == "" {
		cfg.Address = flagAddress
	}
	if cfg.HashKey == "" {
		cfg.HashKey = flagHashKey
	}
	if cfg.CryptoKeyPath == "" {
		cfg.CryptoKeyPath = flagCryptoKeyPath
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = flagReportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = flagPollInterval
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = flagRateLimit
	}
	return nil
}
