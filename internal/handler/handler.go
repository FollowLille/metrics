package handler

import (
	"bytes"
	"fmt"
	"github.com/FollowLille/metrics/internal/compress"
	"go.uber.org/zap"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/FollowLille/metrics/internal/config"
	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
	"github.com/FollowLille/metrics/internal/storage"
)

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
		gz := compress.NewCompressWriter(c.Writer)
		defer gz.Close()
		c.Writer = gz
	}

	// Отправка HTML-страницы в ответе
	c.Data(config.StatusOk, "text/html; charset=utf-8", []byte(html))
}

func UpdateHandler(c *gin.Context, storage *storage.MemStorage) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	if metricName == "" {
		c.String(config.StatusBadRequest, "metric name is empty")
		return
	} else if metricValue == "" {
		c.String(config.StatusBadRequest, "metric value is empty")
		return
	}
	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.String(config.StatusBadRequest, "metric value must be integer")
			return
		}
		storage.UpdateCounter(metricName, value)
		c.String(config.StatusOk, "counter updated")
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.String(config.StatusBadRequest, "metric value must be float")
			return
		}
		storage.UpdateGauge(metricName, value)
		c.String(config.StatusOk, "gauge updated")
	default:
		c.String(config.StatusBadRequest, "metric type must be counter or gauge")
	}
}

func GetValueHandler(c *gin.Context, storage *storage.MemStorage) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	switch metricType {
	case "counter":
		value, exists := storage.GetCounter(metricName)
		if !exists {
			c.String(config.StatusNotFound, "counter with name "+metricName+" not found")
			return
		}
		c.String(config.StatusOk, fmt.Sprintf("%d", value))
	case "gauge":
		value, exists := storage.GetGauge(metricName)
		if !exists {
			c.String(config.StatusNotFound, "gauge with name "+metricName+" not found")
			return
		}
		formattedValue := strconv.FormatFloat(value, 'g', -1, 64)
		c.String(config.StatusOk, formattedValue)
	default:
		c.String(config.StatusBadRequest, "invalid metric type, must be counter or gauge")
	}
}

func UpdateByBodyHandler(c *gin.Context, storage *storage.MemStorage) {
	if c.ContentType() == "application/json" {
		UpdateByJSON(c, storage)
	} else {
		c.String(config.StatusBadRequest, "invalid content type")
	}
}

func UpdateByJSON(c *gin.Context, storage *storage.MemStorage) {
	var metric metrics.Metrics
	c.Header("Content-Type", "application/json")

	// Сохраняем тело запроса для дальнейшего использования
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		c.String(config.StatusBadRequest, "failed to read request body")
		return
	}

	// Восстанавливаем тело запроса для дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&metric); err != nil {
		logger.Log.Error("failed to bind JSON", zap.Error(err))
		c.String(config.StatusBadRequest, "invalid json")
		return
	}
	switch metric.MType {
	case "counter":
		name, value := metric.ID, metric.Delta
		if value == nil {
			c.String(config.StatusBadRequest, "counter value is empty")
			return
		}
		storage.UpdateCounter(name, *value)
		newValue, _ := storage.GetCounter(name)
		metric.Delta = &newValue
		c.JSON(config.StatusOk, metric)
		logger.Log.Info("counter updated", zap.String("counter_name", name), zap.Int64("counter_value", *value))
	case "gauge":
		name, value := metric.ID, metric.Value
		if value == nil {
			c.String(config.StatusBadRequest, "gauge value is empty")
			return
		}
		storage.UpdateGauge(name, *value)
		newValue, _ := storage.GetGauge(name)
		metric.Value = &newValue
		c.JSON(config.StatusOk, metric)
		logger.Log.Info("gauge updated", zap.String("gauge_name", name), zap.Float64("gauge_value", *value))
	default:
		c.String(config.StatusBadRequest, "invalid metric type, must be counter or gauge")
	}
}

func GetValueByBodyHandler(c *gin.Context, storage *storage.MemStorage) {
	if c.ContentType() == "application/json" {
		GetValueByJSON(c, storage)
	} else {
		c.String(config.StatusBadRequest, "invalid content type")
	}
}

func GetValueByJSON(c *gin.Context, storage *storage.MemStorage) {
	var metric metrics.Metrics

	c.Header("Content-Type", "application/json")

	// Сохраняем тело запроса для дальнейшего использования
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("failed to read request body", zap.Error(err))
		c.String(config.StatusBadRequest, "failed to read request body")
		return
	}

	// Восстанавливаем тело запроса для дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.BindJSON(&metric); err != nil {
		logger.Log.Error("failed to bind JSON", zap.Error(err))
		c.Header("Content-Type", "application/json")
		c.String(config.StatusBadRequest, "invalid json")
		return
	}

	name := metric.ID
	switch metric.MType {
	case "counter":
		value, exists := storage.GetCounter(name)
		if !exists {
			c.String(config.StatusNotFound, "counter with name "+name+" not found")
			logger.Log.Info("counter not found", zap.String("counter_name", name))
			return
		}
		metric.Delta = &value
		c.JSON(config.StatusOk, metric)
		logger.Log.Info("counter value", zap.String("counter_name", name), zap.Int64("counter_value", value))
	case "gauge":
		name := metric.ID
		value, exists := storage.GetGauge(name)
		if !exists {
			c.String(config.StatusNotFound, "gauge with name "+name+" not found")
			logger.Log.Info("gauge not found", zap.String("gauge_name", name))
			return
		}
		metric.Value = &value
		c.JSON(config.StatusOk, metric)
		logger.Log.Info("gauge value", zap.String("gauge_name", name), zap.Float64("gauge_value", value))
	default:
		c.String(config.StatusBadRequest, "invalid metric type, must be counter or gauge")
		logger.Log.Info("invalid metric type", zap.String("metric_type", metric.MType))
	}
}
