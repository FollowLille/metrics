package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCalculateHash(t *testing.T) {
	key := []byte("test_key")
	data := []byte("test_data")
	hash := CalculateHash(key, data)

	expectedMac := hmacSHA256(key, data)

	assert.Equal(t, expectedMac, hash, "Hashes should match")
}

func TestVerifyHash(t *testing.T) {
	key := []byte("test_key")
	data := []byte("test_data")
	hash := CalculateHash(key, data)

	assert.True(t, VerifyHash(key, data, []byte(hash)), "Hash verification should succeed")
	assert.False(t, VerifyHash(key, data, []byte("invalid_hash")), "Hash verification should fail")
}

func TestHashMiddleware_Success(t *testing.T) {
	key := []byte("test_key")
	originalBody := []byte("test_body")
	requestHash := CalculateHash(key, originalBody)

	router := gin.Default()
	router.Use(HashMiddleware(key))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(originalBody))
	req.Header.Set("HashSHA256", requestHash)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response status should be OK")
	responseHash := w.Header().Get("HashSHA256")

	assert.NotEmpty(t, responseHash, "Response should contain a hash header")
	assert.Equal(t, CalculateHash(key, w.Body.Bytes()), responseHash, "Response hash should match the expected value")
}

func TestHashMiddleware_Failure(t *testing.T) {
	key := []byte("test_key")
	originalBody := []byte("test_body")
	invalidHash := "invalid_hash"

	router := gin.Default()
	router.Use(HashMiddleware(key))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(originalBody))
	req.Header.Set("HashSHA256", invalidHash)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Response status should be Bad Request")
}

func TestHashMiddleware_NoHashHeader(t *testing.T) {
	key := []byte("test_key")
	originalBody := []byte("test_body")

	router := gin.Default()
	router.Use(HashMiddleware(key))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(originalBody))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response status should be OK even without hash header")
}

// Вспомогательная функция для вычисления HMAC-SHA256 вручную
func hmacSHA256(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
