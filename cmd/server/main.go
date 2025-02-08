// Package main отвечает за запуск сервера
// Он включает в себя функции для парсинга командных флагов и переменных окружения,
// а также настройку логгирования.
package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/compress"
	"github.com/FollowLille/metrics/internal/crypto"
	"github.com/FollowLille/metrics/internal/database"
	"github.com/FollowLille/metrics/internal/handler"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/server"
	"github.com/FollowLille/metrics/internal/storage"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildFlag(buildVersion, buildDate, buildCommit)
	parseFlags()
	metricsStorage := storage.NewMemStorage()

	if err := logger.Initialize(flagLevel); err != nil {
		fmt.Printf("invalid log level: %s", flagLevel)
		return
	}

	// Добавление pprof маршрутов
	go func() {
		pprofRouter := gin.Default()
		pprofRouter.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
		if err := pprofRouter.Run(":6060"); err != nil {
			logger.Log.Error("failed to start pprof router", zap.Error(err))
		}
	}()

	// Запуск сервера
	s := initializeServer(flagAddress, flagCryptoKeyPath)
	router := setupRouter(metricsStorage, s.PrivateKey)

	if err := runServer(s, router, metricsStorage); err != nil {
		panic(err)
	}
}

// setupRouter инициализирует gin и настраивает маршруты
// Принимает хранилище метрик и возвращает gin.Engine
//
// Параметры:
//   - metricsStorage - хранилище метрик
//
// Возвращаемое значение:
//   - *gin.Engine - инициализированный gin.Engine
func setupRouter(metricsStorage *storage.MemStorage, k *rsa.PrivateKey) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.RequestLogger(), logger.ResponseLogger())
	router.Use(crypto.HashMiddleware([]byte(flagHashKey)))
	if flagCryptoKeyPath != "" {
		router.Use(crypto.CryptoDecodeMiddleware(k))
	}
	router.Use(crypto.TrustedSubnetMiddleware(flagTrustedSubnet))
	router.Use(compress.GzipMiddleware(), compress.GzipResponseMiddleware())

	// Маршруты
	router.GET("/", func(c *gin.Context) {
		handler.HomeHandler(c, metricsStorage)
	})

	router.GET("/ping", func(c *gin.Context) {
		handler.PingHandler(c, flagDatabaseAddress)
	})

	router.POST("/update/:type/:name/:value", func(c *gin.Context) {
		handler.UpdateHandler(c, metricsStorage)
	})

	router.POST("/update/", func(c *gin.Context) {
		handler.UpdateByBodyHandler(c, metricsStorage)
	})

	router.POST("/updates", func(c *gin.Context) {
		handler.UpdatesByBodyHandler(c, metricsStorage)
	})

	router.POST("/value/", func(c *gin.Context) {
		handler.GetValueByBodyHandler(c, metricsStorage)
	})

	router.GET("/value/:type/:name", func(c *gin.Context) {
		handler.GetValueHandler(c, metricsStorage)
	})

	return router
}

// initializeServer инициализирует сервер
// Принимает адрес и порт сервера
// Возвращает инициализированный сервер
//
// Параметры:
//   - flags - адрес и порт сервера
//
// Возвращаемое значение:
//   - server.Server - инициализированный сервер
func initializeServer(flags, cryptoKeyPath string) server.Server {
	splittedAddress := strings.Split(flags, ":")
	if len(splittedAddress) != 2 {
		fmt.Printf("invalid address %s, expected host:port", flags)
		os.Exit(1)
	}

	serverAddress := splittedAddress[0]
	serverPort, err := strconv.ParseInt(splittedAddress[1], 10, 64)
	if err != nil {
		fmt.Printf("invalid port: %s", splittedAddress[1])
		os.Exit(1)
	}

	if cryptoKeyPath != "" {
		privateKey, err := crypto.LoadPrivateKey(cryptoKeyPath)
		if err != nil {
			logger.Log.Fatal("failed to load private key", zap.Error(err))
		}
		return server.Server{
			Address:    serverAddress,
			Port:       serverPort,
			PrivateKey: privateKey,
		}
	}

	return server.Server{
		Address: serverAddress,
		Port:    serverPort,
	}
}

// runServer запускает сервер
// Принимает сервер, gin.Engine и хранилище метрик
// Запускает сервер
//
// Параметры:
//   - s - сервер
//   - r - gin.Engine
//   - str - хранилище метрик
//
// Возвращаемое значение:
//   - error - ошибка запуска сервера
func runServer(s server.Server, r *gin.Engine, str *storage.MemStorage) error {
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	logger.Log.Info("starting server", zap.String("address", addr))

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	stopChan := make(chan struct{})

	switch flagStorePlace {
	case "file":
		logger.Log.Info("loading metrics from file")
		storage, file, err := loadMetricsFromFile(str)
		if err != nil {
			return err
		}
		defer file.Close()

		go runPeriodicFileSaver(storage, file, stopChan)
	case "database":
		logger.Log.Info("loading metrics from database")
		database.InitDB(flagDatabaseAddress)
		database.PrepareDB()

		db := database.DB
		if err := database.LoadMetricsFromDatabase(str, db); err != nil {
			return err
		}
		logger.Log.Info("metrics loaded from database")

		go runPeriodicDatabaseSaver(db, stopChan, str)
	default:
		logger.Log.Info("metrics will be stored in memory")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit
	logger.Log.Info("received signal", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Log.Error("failed to shutdown server", zap.Error(err))
		return err
	}

	close(stopChan)
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// loadMetricsFromFile загружает метрики из файла
// Принимает хранилище метрик и возвращает хранилище метрик и файл
//
// Параметры:
//   - str - хранилище метрик
//
// Возвращаемое значение:
//   - *storage.MemStorage - хранилище метрик
//   - *os.File - файл
//   - error - ошибка загрузки метрик из файла
func loadMetricsFromFile(str *storage.MemStorage) (*storage.MemStorage, *os.File, error) {
	if err := os.MkdirAll(flagFilePath, 0755); err != nil {
		logger.Log.Error("can't create directory", zap.Error(err))
		return nil, nil, err
	}

	file, err := os.OpenFile(flagFilePath+"/metrics.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Log.Error("can't open file", zap.Error(err))
		return nil, nil, err
	}

	if flagRestore {
		if err := str.LoadMetricsFromFile(file); err != nil {
			logger.Log.Error("can't load metrics from file", zap.Error(err))
			return nil, nil, err
		}
	}
	return str, file, nil
}

// runPeriodicFileSaver запускает периодическое сохранение метрик в файл
// Принимает хранилище метрик и файл
// Запускает периодическое сохранение метрик в файл
//
// Параметры:
//   - str - хранилище метрик
//   - file - файл
//   - stopChan - канал остановки
func runPeriodicFileSaver(str *storage.MemStorage, file *os.File, stopChan chan struct{}) {
	ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Log.Info("saving metrics to file")
			if err := str.SaveMetricsToFile(file); err != nil {
				logger.Log.Error("can't save metrics to file", zap.Error(err))
			}
		case <-stopChan:
			logger.Log.Info("stop ticker")
			return
		}
	}
}

// runPeriodicDatabaseSaver запускает периодическое сохранение метрик в базу данных
// Принимает базу данных и канал остановки
// Запускает периодическое сохранение метрик в базу данных
//
// Параметры:
//   - db - база данных
//   - stopChan - канал остановки
//   - str - хранилище метрик
func runPeriodicDatabaseSaver(db *sql.DB, stopChan chan struct{}, str *storage.MemStorage) {
	ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Log.Info("saving metrics to database")
			if err := database.SaveMetricsToDatabase(db, str); err != nil {
				logger.Log.Error("can't save metrics to database", zap.Error(err))
			}
		case <-stopChan:
			logger.Log.Info("stop ticker")
			return
		}
	}
}

// printBuildFlag выводит информацию о версии сборки, дате сборки и коммите.
// Если переменные пусты, выводит "N/A".
func printBuildFlag(buildVersion, buildDate, buildCommit string) {
	buildVersion = ifFlagEmpty(buildVersion, "N/A")
	buildDate = ifFlagEmpty(buildDate, "N/A")
	buildCommit = ifFlagEmpty(buildCommit, "N/A")

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
}

// ifFlagEmpty возвращает значение `flag`, если оно не пустое. В противном случае возвращает `alternative`.
func ifFlagEmpty(flag, alternative string) string {
	if flag == "" {
		return alternative
	}
	return flag
}
