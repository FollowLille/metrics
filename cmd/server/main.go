package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/FollowLille/metrics/internal/handler"
	"github.com/FollowLille/metrics/internal/server"
	"github.com/FollowLille/metrics/internal/storage"
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

	// Создаем экземпляр сервера
	s := Init(flagAddress)

	// Запускаем сервер
	err := Run(s, router)
	if err != nil {
		panic(err)
	}
}

func Run(s server.Server, r *gin.Engine) error {
	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	fmt.Printf("server running on: %s", addr)
	err := r.Run(addr)
	return err

}

func Init(flags string) server.Server {
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

	s := server.NewServer()
	s.Address = serverAddress
	s.Port = serverPort
	return *s
}
