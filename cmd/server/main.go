package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/FollowLille/metrics/internal/compress"
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

	// Обработчик стартовой страницы
	router.GET("/", func(context *gin.Context) {
		handler.HomeHandler(context, metricsStorage)
	})

	// Обработчик обновлений
	router.POST("/update", func(c *gin.Context) {
		handler.UpdateByBodyHandler(c, metricsStorage)
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
	fmt.Printf("server running on: %s", addr)

	var err error
	var file *os.File

	if flagRestore {
		file, err = os.OpenFile(flagFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	} else {
		file, err = os.OpenFile(flagFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	}
	if err != nil {
		fmt.Printf("can't open file: %s", flagFilePath)
		logger.Log.Error("can't open file", zap.Error(err))
		return err
	}
	defer file.Close()

	go func() {
		ticker := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)

		for {
			select {
			case <-ticker.C:
				err := s.SaveMetricsToFile(str, file)
				if err != nil {
					fmt.Printf("can't save metrics to file: %s", flagFilePath)
					logger.Log.Error("can't save metrics to file", zap.Error(err))
				}
			}
		}
	}()

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
