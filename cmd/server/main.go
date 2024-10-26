package main

import (
	"database/sql"
	"fmt"
	"github.com/FollowLille/metrics/internal/crypto"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/compress"
	"github.com/FollowLille/metrics/internal/database"
	"github.com/FollowLille/metrics/internal/handler"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/server"
	"github.com/FollowLille/metrics/internal/storage"
)

func main() {
	// Инициализация хранилища и роутера
	parseFlags()
	metricsStorage := storage.NewMemStorage()

	// Инициализация логгера
	if err := logger.Initialize(flagLevel); err != nil {
		fmt.Printf("invalid log level: %s", flagLevel)
		os.Exit(1)
	}
	// Инициализация роутера с восстановлением
	router := gin.New()
	router.Use(gin.Recovery())

	// Инициализация обработчиков
	router.Use(logger.RequestLogger()).Use(logger.ResponseLogger())

	// Инициализация сжатия
	router.Use(compress.GzipMiddleware()).Use(compress.GzipResponseMiddleware())

	// Инициализация хэша
	if flagHashKey != "" {
		router.Use(crypto.HashMiddleware([]byte(flagHashKey)))
	}

	// Обработчик стартовой страницы
	router.GET("/", func(context *gin.Context) {
		handler.HomeHandler(context, metricsStorage)
	})

	// Обработчик пинга к базе
	router.GET("/ping", func(c *gin.Context) {
		handler.PingHandler(c, flagDatabaseAddress)
	})

	// Обработчик обновлений
	router.POST("/update", func(c *gin.Context) {
		handler.UpdateByBodyHandler(c, metricsStorage)
	})

	router.POST("/updates", func(c *gin.Context) {
		handler.UpdatesByBodyHandler(c, metricsStorage)
	})

	router.POST("/update/:type/:name/:value", func(c *gin.Context) {
		handler.UpdateHandler(c, metricsStorage)
	})

	// Обработчик получения метрик
	router.POST("/value", func(c *gin.Context) {
		handler.GetValueByBodyHandler(c, metricsStorage)
	})

	router.GET("/value/:type/:name", func(c *gin.Context) {
		handler.GetValueHandler(c, metricsStorage)
	})

	// Создаем экземпляр сервера
	s := Initialize(flagAddress)

	// Запускаем сервер
	err := Run(s, router, metricsStorage)
	if err != nil {
		panic(err)
	}
}

func Run(s server.Server, r *gin.Engine, str *storage.MemStorage) error {
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	logger.Log.Info("starting server", zap.String("address", addr))
	var err error
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
		err = database.LoadMetricsFromDatabase(str, db)
		if err != nil {
			return err
		}
		logger.Log.Info("metrics loaded from database")

		go runPeriodicDatabaseSaver(db, stopChan, str)
	default:
		logger.Log.Info("metrics will be stored in memory")
	}

	err = r.Run(addr)
	return err
}

func Initialize(flags string) server.Server {
	splitedAddress := strings.Split(flags, ":")
	if len(splitedAddress) != 2 {
		fmt.Printf("invalid address %s, expected host:port", flags)
		os.Exit(1)
	}

	serverAddress := splitedAddress[0]
	serverPort, err := strconv.ParseInt(splitedAddress[1], 10, 64)
	if err != nil {
		fmt.Printf("invalid port: %s", splitedAddress[1])
		os.Exit(1)
	}

	if err := logger.Initialize(flagLevel); err != nil {
		fmt.Printf("invalid log level: %s", flagLevel)
		os.Exit(1)
	}
	s := server.NewServer()
	s.Address = serverAddress
	s.Port = serverPort
	return *s
}

func loadMetricsFromFile(str *storage.MemStorage) (*storage.MemStorage, *os.File, error) {
	var err error
	var file *os.File

	if err := os.MkdirAll(flagFilePath, 0755); err != nil {
		logger.Log.Error("can't create dir", zap.Error(err))
		return nil, nil, err
	}

	file, err = os.OpenFile(flagFilePath+"/metrics.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Log.Error("can't open file", zap.Error(err))
		return nil, nil, err
	}

	if flagRestore {
		err = str.LoadMetricsFromFile(file)
		if err != nil {
			logger.Log.Error("can't load metrics from file", zap.Error(err))
			return nil, nil, err
		}
	}
	return str, file, nil
}

func runPeriodicFileSaver(str *storage.MemStorage, file *os.File, stopChan chan struct{}) {
	ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
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
