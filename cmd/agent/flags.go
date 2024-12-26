// Package main для запуска агента
// Данная часть содержит информацию об использемых флагах
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/logger"
)

// Флаги
var (
	flagAddress        string // адрес для прослушивания
	flagHashKey        string // ключ хэша
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
//	     -hash-key=secret
//			-report-interval=10
//			-poll-interval=2
//			-rate-limit=4
//
// После парсинга флагов, информация о них логируется с использованием zap.
func parseFlags() error {
	pflag.StringVarP(&flagAddress, "address", "a", "localhost:8080", "address")
	pflag.StringVarP(&flagHashKey, "hash-key", "k", "", "hash key")
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

	logger.Log.Info("Flags",
		zap.String("address", flagAddress),
		zap.String("hash-key", flagHashKey),
		zap.Int64("report-interval", flagReportInterval),
		zap.Int64("poll-interval", flagPollInterval),
		zap.Int64("rate-limit", flagRateLimit),
	)
	return nil

}
