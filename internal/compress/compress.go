// Package compress реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/logger"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	*responseWriter
	zw *gzip.Writer
}

// NewCompressWriter создает новый gzip.Writer
// Если произошла ошибка, то логируем ее и возвращаем nil
// иначе возвращаем gzip.Writer
//
// Параметры:
//   - w - http.ResponseWriter
//
// Возвращаемое значение:
//   - *compressWriter
func NewCompressWriter(w *responseWriter) *compressWriter {
	zw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		logger.Log.Error("failed to create gzip writer", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	return &compressWriter{
		responseWriter: w,
		zw:             zw,
	}
}

// Header возвращает http.Header
//
// Возвращаемое значение:
//   - http.Header
func (c *compressWriter) Header() http.Header {
	return c.responseWriter.Header()
}

// WriteHeader записывает код статуса в http.Header
//
// Параметры:
//   - p массив символов
//
// Возвращаемое значение:
//   - int
//   - error
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader записывает код статуса в http.Header
//
// Параметры:
//   - statusCode int
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.responseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.responseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает новый gzip.Reader
// Если произошла ошибка, то логируем ее и возвращаем nil
// иначе возвращаем gzip.Reader
//
// Параметры:
//   - r io.ReadCloser
//
// Возвращаемое значение:
//   - *compressReader
//   - error
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read декомпрессирует данные из буфера
//
// Параметры:
//   - p []byte
//
// Возвращаемое значение:
//   - int
//   - error
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает gzip.Reader и досылает все данные из буфера.
//
// Возвращаемое значение:
//   - error
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware обрабатывает запросы
// Если в заголовках в Content-Encoding содержится gzip, то создаем gzip.Writer и обрабатывает запрос.
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				logger.Log.Error("failed to create gzip reader", zap.Error(err))
				return
			}
			defer gz.Close()
			body, err := io.ReadAll(gz)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				logger.Log.Error("failed to read gzip body", zap.Error(err))
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()
	}
}

// GzipResponseMiddleware обрабатывает ответы
// Если в заголовках в Accept-Encoding содержится gzip, то создаем gzip.Writer
// и декомпрессируем тело ответа.
func GzipResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := c.Writer

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
				c.Header("Content-Encoding", "gzip")
				wr := NewResponseWriter(w)
				gz := NewCompressWriter(wr)
				defer func() {
					gz.Flush()
					gz.Close()
				}()
				c.Writer = gz
			}
		}
		c.Next()
	}
}

// responseWriter реализует интерфейс gin.ResponseWriter и позволяет прозрачно для сервера
// записывать данные в буфер
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// NewResponseWriter создает новый responseWriter
//
// Параметры:
//   - gin.ResponseWriter
//
// Возвращаемое значение:
//   - *responseWriter
func NewResponseWriter(w gin.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer([]byte{}),
	}
}

// Write записывает данные в буфер
//
// Параметры:
//   - p []byte
//
// Возвращаемое значение:
//   - int
//   - error
func (w *responseWriter) Write(p []byte) (int, error) {
	n, err := w.body.Write(p)
	if err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

// GetBody возвращает содержимое буфера
func (w *responseWriter) GetBody() []byte {
	return w.body.Bytes()
}
