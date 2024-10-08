package logger

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log = zap.NewNop()

// Инициализация логгера
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

// Логирование запроса
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

// Логирование ответа
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
