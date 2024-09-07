package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/FollowLille/metrics/internal/handler"
	"github.com/FollowLille/metrics/internal/server"
	"github.com/FollowLille/metrics/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализация хранилища и роутера
	parseFlags()
	metricsStorage := storage.NewMemStorage()

	router := gin.Default()

	// Обработчик стартовой страницы
	router.GET("/", func(context *gin.Context) {
		handler.HomeHandler(context, metricsStorage)
	})

	// Обработчик обновлений
	router.POST("/update/:type/:name/:value", func(c *gin.Context) {
		handler.UpdateHandler(c, metricsStorage)
	})

	// Обработчик получения метрик
	router.GET("/value/:type/:name", func(c *gin.Context) {
		handler.GetValueHandler(c, metricsStorage)
	})

	// Запуск HTTP-сервера
	s := server.NewServer()
	splitedAddress := strings.Split(flagAddress, ":")
	serverAddress := splitedAddress[0]
	serverPort, err := strconv.ParseInt(splitedAddress[1], 10, 64)
	if err != nil {
		fmt.Printf("invalid address: %s", flagAddress)
		os.Exit(1)
	}
	s.Address = serverAddress
	s.Port = serverPort
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	fmt.Println("server running on:", addr)
	err = router.Run(addr)

	if err != nil {
		panic(err)
	}
}
