package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"

	"github.com/FollowLille/metrics/internal/logger"
)

func CalculateHash(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHash(key, data, hash []byte) bool {
	expectedHash := CalculateHash(key, data)
	return hmac.Equal(hash, []byte(expectedHash))
}

// Была развилка сделать отдельно в каждом шаге хеширование или переопределить обработчик запроса, чтобы можно было получить доступ к телу после всех обработок

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.body.Write(p)
	return w.ResponseWriter.Write(p)
}

func HashMiddleware(key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Проверка хеша запроса
		hash := c.Request.Header.Get("HashSHA256")
		if hash == "" {
			logger.Log.Info("Hash not found in request header")
			c.Next()
		} else {
			logger.Log.Info("Hash found in request header")
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Log.Error("Failed to read request body", zap.Error(err))
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			// Восстанавливаем тело запроса для последующей обработки
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			// Проверка хэша
			if !VerifyHash(key, body, []byte(hash)) {
				logger.Log.Error("Hash verification failed")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			w := &responseWriter{c.Writer, bytes.NewBuffer([]byte{})}
			c.Writer = w

			// Выполнение следующего обработчика
			c.Next()

			// Добавляем хэш в заголовок ответа
			responseHash := CalculateHash(key, w.body.Bytes())
			c.Writer.Header().Set("HashSHA256", responseHash)
		}
	}
}
