package main

import (
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
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

func main() {
	// Инициализация хранилища и логгера
	parseFlags()
	metricsStorage := storage.NewMemStorage()

	if err := logger.Initialize(flagLevel); err != nil {
		fmt.Printf("invalid log level: %s", flagLevel)
		os.Exit(1)
	}

	// Добавление pprof маршрутов
	go func() {
		pprofRouter := gin.Default()
		pprofRouter.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
		if err := pprofRouter.Run(":6060"); err != nil {
			logger.Log.Error("failed to start pprof router", zap.Error(err))
		}
	}()

	// Инициализация роутера
	router := setupRouter(metricsStorage)

	// Создание и запуск сервера
	s := initializeServer(flagAddress)
	if err := runServer(s, router, metricsStorage); err != nil {
		panic(err)
	}
}

func setupRouter(metricsStorage *storage.MemStorage) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.RequestLogger(), logger.ResponseLogger())
	router.Use(crypto.HashMiddleware([]byte(flagHashKey)))
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

func initializeServer(flags string) server.Server {
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

	return server.Server{
		Address: serverAddress,
		Port:    serverPort,
	}
}

func runServer(s server.Server, r *gin.Engine, str *storage.MemStorage) error {
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	logger.Log.Info("starting server", zap.String("address", addr))

	stopChan := make(chan struct{})

	switch flagStorePlace {
	case "file":
		logger.Log.Info("loading metrics from file")
		str, file, err := loadMetricsFromFile(str)
		if err != nil {
			return err
		}
		defer file.Close()

		go runPeriodicFileSaver(str, file, stopChan)
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

	return r.Run(addr)
}

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
