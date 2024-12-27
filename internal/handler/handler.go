// Package handler содержит функции для обработки HTTP-запросов
package handler

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/compress"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
	"github.com/FollowLille/metrics/internal/storage"
)

// HomeHandler обрабатывает GET-запрос на "/"
// Принимает хранилище метрик и возвращает HTML-страницу
//
// Параметры:
//   - c - gin.Context
//   - s - хранилище метрик
func HomeHandler(c *gin.Context, s *storage.MemStorage) {
	// Получение всех метрик
	gauges := s.GetAllGauges()
	counters := s.GetAllCounters()

	// Формирование HTML-страницы
	html := "<!DOCTYPE html><html><head><title>Metrics</title></head><body>"
	html += "<h1>Metrics</h1>"

	html += "<h2>Counters</h2><ul>"
	for name, value := range counters {
		html += fmt.Sprintf("<li>%s: %d</li>", name, value)
	}
	html += "</ul>"

	html += "<h2>Gauges</h2><ul>"
	for name, value := range gauges {
		html += fmt.Sprintf("<li>%s: %.2f</li>", name, value)
	}
	html += "</ul>"

	html += "</body></html>"

	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		c.Header("Content-Encoding", "gzip")
		wr := compress.NewResponseWriter(c.Writer)
		gz := compress.NewCompressWriter(wr)
		defer gz.Close()
		c.Writer = gz
	}

	// Отправка HTML-страницы в ответе
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// UpdateHandler обрабатывает PUT-запрос на "/update/{type}/{name}/{value}"
// Принимает хранилище метрик и обновляет значение метрики
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func UpdateHandler(c *gin.Context, storage *storage.MemStorage) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	if metricName == "" {
		c.String(http.StatusBadRequest, "metric name is empty")
		return
	} else if metricValue == "" {
		c.String(http.StatusBadRequest, "metric value is empty")
		return
	}
	switch metricType {
	case metrics.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "metric value must be integer")
			return
		}
		storage.UpdateCounter(metricName, value)
		c.String(http.StatusOK, "counter updated")
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "metric value must be float")
			return
		}
		storage.UpdateGauge(metricName, value)
		c.String(http.StatusOK, "gauge updated")
	default:
		c.String(http.StatusBadRequest, "metric type must be counter or gauge")
	}
}

// GetValueHandler обрабатывает GET-запрос на "/value/{type}/{name}"
// Принимает хранилище метрик и возвращает значение метрики
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func GetValueHandler(c *gin.Context, storage *storage.MemStorage) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	switch metricType {
	case metrics.Counter:
		value, exists := storage.GetCounter(metricName)
		if !exists {
			c.String(http.StatusNotFound, "counter with name "+metricName+" not found")
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%d", value))
	case metrics.Gauge:
		value, exists := storage.GetGauge(metricName)
		if !exists {
			c.String(http.StatusNotFound, "gauge with name "+metricName+" not found")
			return
		}
		formattedValue := strconv.FormatFloat(value, 'g', -1, 64)
		c.String(http.StatusOK, formattedValue)
	default:
		c.String(http.StatusBadRequest, "invalid metric type, must be counter or gauge")
	}
}

// UpdateByBodyHandler обрабатывает POST-запрос на "/update"
// Принимает хранилище метрик и обновляет значения метрик
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func UpdateByBodyHandler(c *gin.Context, storage *storage.MemStorage) {
	if c.ContentType() == "application/json" {
		UpdateByJSON(c, storage)
	} else {
		c.String(http.StatusBadRequest, "invalid content type")
	}
}

// UpdatesByBodyHandler обрабатывает POST-запрос на "/updates"
// Принимает хранилище метрик и обновляет значения метрик
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func UpdatesByBodyHandler(c *gin.Context, storage *storage.MemStorage) {
	if c.ContentType() == "application/json" {
		UpdatesByJSON(c, storage)
	} else {
		c.String(http.StatusBadRequest, "invalid content type")
	}
}

// UpdateByJSON обрабатывает POST-запрос на "/update"
// Принимает хранилище метрик и обновляет значения метрик
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func UpdateByJSON(c *gin.Context, storage *storage.MemStorage) {
	var metric metrics.Metrics

	// Сохраняем тело запроса для дальнейшего использования
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		c.String(http.StatusBadRequest, "failed to read request body")
		return
	}

	// Восстанавливаем тело запроса для дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&metric); err != nil {
		logger.Log.Error("failed to bind JSON", zap.Error(err))
		c.String(http.StatusBadRequest, "invalid json")
		return
	}
	switch metric.MType {
	case metrics.Counter:
		name, value := metric.ID, metric.Delta
		if value == nil {
			c.String(http.StatusBadRequest, "counter value is empty")
			return
		}
		storage.UpdateCounter(name, *value)
		newValue, _ := storage.GetCounter(name)
		metric.Delta = &newValue
		c.JSON(http.StatusOK, metric)
		logger.Log.Info("counter updated", zap.String("counter_name", name), zap.Int64("counter_value", *value))
	case metrics.Gauge:
		name, value := metric.ID, metric.Value
		if value == nil {
			c.String(http.StatusBadRequest, "gauge value is empty")
			return
		}
		storage.UpdateGauge(name, *value)
		newValue, _ := storage.GetGauge(name)
		metric.Value = &newValue
		c.JSON(http.StatusOK, metric)
		logger.Log.Info("gauge updated", zap.String("gauge_name", name), zap.Float64("gauge_value", *value))
	default:
		c.String(http.StatusBadRequest, "invalid metric type, must be counter or gauge")
	}
}

// UpdatesByJSON обрабатывает POST-запрос на "/updates"
// Принимает хранилище метрик и обновляет значения метрик
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func UpdatesByJSON(c *gin.Context, storage *storage.MemStorage) {
	var metricsBatch []metrics.Metrics

	// Сохраняем тело запроса для дальнейшего использования
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		c.String(http.StatusBadRequest, "failed to read request body")
		return
	}

	// Восстанавливаем тело запроса для дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&metricsBatch); err != nil {
		logger.Log.Error("failed to bind JSON", zap.Error(err))
		c.String(http.StatusBadRequest, "invalid json")
		return
	}

	for _, metric := range metricsBatch {
		switch metric.MType {
		case metrics.Counter:
			name, value := metric.ID, metric.Delta
			if value == nil {
				c.String(http.StatusBadRequest, "counter value is empty")
				return
			}
			storage.UpdateCounter(name, *value)
			logger.Log.Info("counter updated", zap.String("counter_name", name), zap.Int64("counter_value", *value))
		case metrics.Gauge:
			name, value := metric.ID, metric.Value
			if value == nil {
				c.String(http.StatusBadRequest, "gauge value is empty")
				return
			}
			storage.UpdateGauge(name, *value)
			logger.Log.Info("gauge updated", zap.String("gauge_name", name), zap.Float64("gauge_value", *value))
		default:
			c.String(http.StatusBadRequest, "invalid metric type, must be counter or gauge")
		}
	}
	c.JSON(http.StatusOK, metricsBatch)
}

// GetValueByBodyHandler обрабатывает GET-запрос на "/value"
// Принимает хранилище метрик и возвращает значение метрики
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func GetValueByBodyHandler(c *gin.Context, storage *storage.MemStorage) {
	if c.ContentType() == "application/json" {
		GetValueByJSON(c, storage)
	} else {
		c.String(http.StatusBadRequest, "invalid content type")
	}
}

// GetValueByJSON обрабатывает GET-запрос на "/value"
// Принимает хранилище метрик и возвращает значение метрики
//
// Параметры:
//   - c - gin.Context
//   - storage - хранилище метрик
func GetValueByJSON(c *gin.Context, storage *storage.MemStorage) {
	var metric metrics.Metrics

	// Сохраняем тело запроса для дальнейшего использования
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		c.String(http.StatusBadRequest, "failed to read request body")
		return
	}

	// Восстанавливаем тело запроса для дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.BindJSON(&metric); err != nil {
		logger.Log.Error("failed to bind JSON", zap.Error(err))
		c.String(http.StatusBadRequest, "invalid json")
		return
	}
	logger.Log.Info("received metric", zap.Any("metric", metric))
	name := metric.ID
	switch metric.MType {
	case metrics.Counter:
		value, exists := storage.GetCounter(name)
		logger.Log.Info("counter value", zap.String("counter_name", name), zap.Int64("counter_value", value))
		if !exists {
			c.String(http.StatusNotFound, "counter with name "+name+" not found")
			logger.Log.Info("counter not found", zap.String("counter_name", name))
			return
		}

		metric.Delta = &value
		c.JSON(http.StatusOK, metric)
		logger.Log.Info("counter value", zap.String("counter_name", name), zap.Int64("counter_value", value))
	case metrics.Gauge:
		name := metric.ID
		value, exists := storage.GetGauge(name)
		logger.Log.Info("gauge value", zap.String("gauge_name", name), zap.Float64("gauge_value", value))
		if !exists {
			c.String(http.StatusNotFound, "gauge with name "+name+" not found")
			logger.Log.Info("gauge not found", zap.String("gauge_name", name))
			return
		}
		metric.Value = &value
		c.JSON(http.StatusOK, metric)
	default:
		c.String(http.StatusBadRequest, "invalid metric type, must be counter or gauge")
		logger.Log.Info("invalid metric type", zap.String("metric_type", metric.MType))
	}
}

// PingHandler обрабатывает GET-запрос на "/ping"
// Принимает адрес базы данных и возвращает "pong"
//
// Параметры:
//   - c - gin.Context
//   - adr - адрес базы данных
func PingHandler(c *gin.Context, adr string) {
	db, err := sql.Open("postgres", adr)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to connect to db")
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		c.String(http.StatusInternalServerError, "failed to ping db")
		return
	}
	c.String(http.StatusOK, "pong")
}
