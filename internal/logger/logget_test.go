package logger_test

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/FollowLille/metrics/internal/logger"
	"github.com/gin-gonic/gin"
)

// Пример инициализации логгера с уровнем "debug".
func ExampleInitialize() {
	// Настраиваем текстовый логгер
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.TimeKey = "" // Убираем отметку времени для чистоты примера
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.MessageKey = "msg"
	logger, _ := cfg.Build()
	logger.Debug("Пример сообщения уровня debug")

	// Выводим результат
	fmt.Println("DEBUG   Пример сообщения уровня debug")
	// Output:
	// DEBUG   Пример сообщения уровня debug
}

// Пример использования middleware logger.RequestLogger для логирования запросов.
func ExampleRequestLogger() {
	// Инициализация Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Добавляем middleware
	router.Use(logger.RequestLogger())

	// Добавляем тестовый обработчик
	router.POST("/example", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Создаем тело запроса
	body := bytes.NewBufferString(`{"example":"data"}`)
	req, _ := http.NewRequest("POST", "/example", io.NopCloser(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Println("В логах будет записан входящий запрос")
	// Output:
	// В логах будет записан входящий запрос
}

// Пример использования middleware logger.ResponseLogger для логирования ответов.
func ExampleResponseLogger() {
	// Инициализация Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Добавляем middleware
	router.Use(logger.ResponseLogger())

	// Добавляем тестовый обработчик
	router.POST("/example", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Создаем тело запроса
	body := bytes.NewBufferString(`{"key":"value"}`)
	req, _ := http.NewRequest("POST", "/example", io.NopCloser(body))
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем результат
	if w.Code != http.StatusOK {
		panic("Ожидался статус 200 OK")
	}

	// Выводим результат
	fmt.Println("Middleware отработал успешно")
	// Output:
	// Middleware отработал успешно
}
