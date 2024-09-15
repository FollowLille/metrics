package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/FollowLille/metrics/internal/config"
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
