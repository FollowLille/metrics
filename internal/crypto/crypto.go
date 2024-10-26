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

// CalculateHash computes HMAC SHA256 hash
func CalculateHash(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHash verifies HMAC SHA256 hash
func VerifyHash(key, data, hash []byte) bool {
	expectedHash := CalculateHash(key, data)
	return hmac.Equal(hash, []byte(expectedHash))
}

// hashResponseWriter wraps gin.ResponseWriter to capture response body
type hashResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewHashResponseWriter creates a new hashResponseWriter
func NewHashResponseWriter(w gin.ResponseWriter) *hashResponseWriter {
	return &hashResponseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer([]byte{}),
	}
}

// Write captures the response body
func (w *hashResponseWriter) Write(p []byte) (int, error) {
	n, err := w.body.Write(p)
	if err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

// GetBody returns the captured response body
func (w *hashResponseWriter) GetBody() []byte {
	return w.body.Bytes()
}

func HashMiddleware(key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Writer.Header().Set("HashSHA256", responseHash)
	}
}
