// Package logger содержит функции для логирования
package logger

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log = zap.NewNop()

// Initialize инициализирует логгер
// Принимает уровень логирования и возвращает ошибку, если она возникла
//
// Параметры:
//   - level - уровень логирования
//
// Возвращаемое значение:
//   - error - ошибка
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

// RequestLogger логирует запросы
// В лог попадают следующие параметры:
//   - method - метод
//   - path - путь
//   - duration - время выполнения
//   - body - тело запроса
//   - headers - заголовки запроса
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			Log.Error("failed to read request body", zap.Error(err))
			c.Next() // передаем обработку дальше
			return
		}

		headers := c.Request.Header
		headerMap := make(map[string]string)
		for key, values := range headers {
			headerMap[key] = values[0] // Логируем первый элемент (если несколько значений)
		}

		Log.Info("got incoming HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("duration", time.Since(start)),
			zap.ByteString("body", bodyBytes),
			zap.Any("headers", headerMap),
		)

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

// ResponseLogger логирует ответы
// В лог попадают следующие параметры:
//   - status - статус
//   - response_size - размер ответа
//   - body - тело ответа
//   - headers - заголовки
func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			Log.Error("failed to read request body", zap.Error(err))
			c.Next() // передаем обработку дальше
			return
		}

		headers := c.Request.Header
		headerMap := make(map[string]string)
		for key, values := range headers {
			headerMap[key] = values[0] // Логируем первый элемент (если несколько значений)
		}
		Log.Info("HTTP response",
			zap.Int("status", statusCode),
			zap.Int("response_size", responseSize),
			zap.ByteString("body", bodyBytes),
			zap.Any("headers", headerMap),
		)
	}
}
