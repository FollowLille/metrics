// Package crypto содержит функции для шифрования данных
package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/logger"
)

// CalculateHash вычисляет хеш SHA256
// Принимает ключ и данные и возвращает хеш в виде строки
//
// Параметры:
//   - key - ключ
//   - data - данные к шифрованию
//
// Возвращаемое значение:
//   - хеш в виде строки
func CalculateHash(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHash проверяет хеш SHA256
// Принимает ключ, данные и хеш и возвращает true, если хеш совпадает
//
// Параметры:
//   - key - ключ
//   - data - данные к шифрованию
//   - hash - хеш
//
// Возвращаемое значение:
//   - true, если хеш совпадает
func VerifyHash(key, data, hash []byte) bool {
	expectedHash := CalculateHash(key, data)
	return hmac.Equal(hash, []byte(expectedHash))
}

type hashResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewHashResponseWriter создает новый gin.ResponseWriter
// и возвращает его в виде *hashResponseWriter
func NewHashResponseWriter(w gin.ResponseWriter) *hashResponseWriter {
	return &hashResponseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer([]byte{}),
	}
}

// Write записывает данные в body
//
// Параметры:
//   - p массив символов
//
// Возвращаемое значение:
//   - int
//   - error
func (w *hashResponseWriter) Write(p []byte) (int, error) {
	n, err := w.body.Write(p)
	if err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

// GetBody возвращает содержимое body
func (w *hashResponseWriter) GetBody() []byte {
	return w.body.Bytes()
}

// HashMiddleware добавляет хеш в запрос
// Принимает ключ и возвращает gin.HandlerFunc
func HashMiddleware(key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if string(key) == "" {
			c.Next()
			return
		}
		hash := c.Request.Header.Get("HashSHA256")
		if hash == "" {
			logger.Log.Info("Hash not found in request header")
			// Просто продолжаем выполнение следующих обработчиков
			c.Next()
			return
		}

		logger.Log.Info("Hash found in request header")
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Failed to read request body", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if !VerifyHash(key, body, []byte(hash)) {
			logger.Log.Error("Hash verification failed")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		w := NewHashResponseWriter(c.Writer)
		c.Writer = w

		c.Next()

		originalBody := w.GetBody()
		responseHash := CalculateHash(key, originalBody)
		c.Header("HashSHA256", responseHash)
	}
}
