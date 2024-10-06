package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/config"
	"github.com/FollowLille/metrics/internal/logger"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
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

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(config.StatusBadRequest)
				logger.Log.Error("failed to create gzip reader", zap.Error(err))
				return
			}
			defer gz.Close()
			body, err := io.ReadAll(gz)
			if err != nil {
				c.AbortWithStatus(config.StatusBadRequest)
				logger.Log.Error("failed to read gzip body", zap.Error(err))
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()
	}
}

func GzipResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := c.Writer

		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			contentType := c.GetHeader("Content-Type")
			fmt.Println("content type: ", contentType)
			if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/json") {
				c.Header("Content-Encoding", "gzip")
				gz := NewCompressWriter(w)
				defer gz.Close()
				c.Writer = gz
			}
		}
		c.Next()
	}
}
