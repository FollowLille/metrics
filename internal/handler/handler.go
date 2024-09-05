package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FollowLille/metrics/internal/storage"
	"github.com/gin-gonic/gin"
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
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func UpdateHandler(c *gin.Context, storage *storage.MemStorage) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	if metricName == "" {
		c.String(http.StatusBadRequest, "Не заполнено имя метрики")
	} else if metricValue == "" {
		c.String(http.StatusBadRequest, "Не заполнено значение метрики")
	}
	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Значение метрики должно быть целым числом")
			return
		}
		storage.UpdateCounter(metricName, value)
		c.String(http.StatusOK, "Counter обновлен")
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Значение метрики должно быть числом с плавающей точкой")
			return
		}
		storage.UpdateGauge(metricName, value)
		c.String(http.StatusOK, "Gauge обновлен")
	default:
		c.String(http.StatusBadRequest, "Тип метрики должен быть counter или gauge")
	}
}

func GetValueHandler(c *gin.Context, storage *storage.MemStorage) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	switch metricType {
	case "counter":
		value, exists := storage.GetCounter(metricName)
		if !exists {
			c.String(http.StatusNotFound, "Counter с именем "+metricName+" не найден")
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%d", value))
	case "gauge":
		value, exists := storage.GetGauge(metricName)
		if !exists {
			c.String(http.StatusNotFound, "Gauge с именем "+metricName+" не найден")
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%f", value))
	default:
		c.String(http.StatusBadRequest, "Некорректное имя метрики, требуется counter или gauge")
	}
}
