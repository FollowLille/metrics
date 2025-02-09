// Package crypto содержит функции для шифрования данных
package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"os"

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
		// Всегда создаем NewHashResponseWriter, потому что он будет переиспользван потом
		w := NewHashResponseWriter(c.Writer)
		c.Writer = w

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

		c.Next()

		originalBody := w.GetBody()
		responseHash := CalculateHash(key, originalBody)
		c.Header("HashSHA256", responseHash)
	}
}

// LoadPrivateKey загружает RSA-ключ из файла
// Принимает путь к файлу и возвращает RSA-ключ
//
// Параметры:
//   - filePath - путь к файлу
//
// Возвращаемое значение:
//   - RSA-ключ
//   - error
func LoadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	decoded, _ := pem.Decode(keyData)
	if decoded == nil || decoded.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key file")
	}

	return x509.ParsePKCS1PrivateKey(decoded.Bytes)
}

// LoadPublicKey загружает RSA-ключ из файла
// Принимает путь к файлу и возвращает RSA-ключ
//
// Параметры:
//   - filePath - путь к файлу
//
// Возвращаемое значение:
//   - RSA-ключ
//   - error
func LoadPublicKey(filePath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	decoded, _ := pem.Decode(keyData)
	if decoded == nil || decoded.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key file")
	}
	key, err := x509.ParsePKIXPublicKey(decoded.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key")
	}
	return pubKey, nil
}

// Encrypt шифрует данные
// Принимает RSA-ключ и возвращает зашифрованные данные
//
// Параметры:
//   - publicKey - RSA-ключ
//   - data - данные
//
// Возвращаемое значение:
//   - зашифрованные данные
//   - error
func Encrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

// Decrypt дешифрует данные
// Принимает RSA-ключ и возвращает расшифрованные данные
//
// Параметры:
//   - privateKey - RSA-ключ
//   - data - данные
//
// Возвращаемое значение:
//   - расшифрованные данные
//   - error
func Decrypt(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
}

func CryptoDecodeMiddleware(privateKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Failed to read request body", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		decryptedData, err := Decrypt(privateKey, body)
		if err != nil {
			logger.Log.Error("Failed to decrypt data", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedData))
		c.Next()
	}
}

// TrustedSubnetMiddleware проверяет, что IP-адрес принадлежит подсети trustedSubnet
// Если IP-адрес не принадлежит подсети, то возвращается 403 ошибка
// Принимает CIDR-подсеть и возвращает gin.HandlerFunc
//
// Параметры:
//   - trustedSubnet - CIDR-подсеть
//
// Возвращаемое значение:
//   - gin.HandlerFunc
func TrustedSubnetMiddleware(trustedSubnet string) gin.HandlerFunc {
	if trustedSubnet == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logger.Log.Error("Failed to parse trusted subnet", zap.Error(err))
	}
	return func(c *gin.Context) {
		ipStr := c.Request.Header.Get("X-Real-IP")
		if ipStr == "" {
			logger.Log.Warn("X-Real-IP header not found")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil || !subnet.Contains(ip) {
			logger.Log.Warn("Invalid X-Real-IP header")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

// DecodeGRPCRequest декодирует запрос в структуру
// Принимает структуру и RSA-ключ и возвращает структуру
//
// Параметры:
//   - req - структура
//   - privateKey - RSA-ключ
//
// Возвращаемое значение:
//   - структура
//   - error
func DecodeGRPCRequest(req interface{}, privateKey *rsa.PrivateKey) (interface{}, error) {
	if privateKey == nil {
		return nil, errors.New("private key is nil")
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to marshal request", zap.Error(err))
		return nil, err
	}

	decryptedData, err := Decrypt(privateKey, reqBytes)
	if err != nil {
		logger.Log.Error("Failed to decrypt request", zap.Error(err))
		return nil, err
	}

	err = json.Unmarshal(decryptedData, req)
	if err != nil {
		logger.Log.Error("Failed to unmarshal request", zap.Error(err))
		return nil, err
	}

	return req, nil
}
