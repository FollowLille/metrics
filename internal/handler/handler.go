package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/FollowLille/metrics/internal/storage"
)

func HomeHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Страница не найдена", http.StatusNotFound)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, storage *storage.MemStorage) {
	if r.Method != http.MethodPost {
		http.Error(w, "Можно использовать только метод Post", http.StatusMethodNotAllowed)
		return
	}
	fullPath, err := url.Parse(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metricType, metricName, metricValue, err := parseAndValidatePath(fullPath.Path, w)
	if err != nil {
		return
	}

	if metricType == "counter" {
		valueInt, _ := strconv.ParseInt(metricValue, 10, 64)
		storage.UpdateCounter(metricName, valueInt)
	} else if metricType == "gauge" {
		valueFloat, _ := strconv.ParseFloat(metricValue, 64)
		storage.UpdateGauge(metricName, valueFloat)
	}
}

func parseAndValidatePath(path string, w http.ResponseWriter) (string, string, string, error) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) <= 2 {
		http.Error(w, "Некорректный запрос, пожалуйста, попробуйте ещё раз", http.StatusNotFound)
		return "", "", "", fmt.Errorf("некорректный запрос: %s", path)
	}
	if len(segments) != 4 {
		http.Error(w, "Некорректный запрос, пожалуйста, попробуйте ещё раз", http.StatusBadRequest)
		return "", "", "", fmt.Errorf("некорректный запрос: %s", path)
	}
	if segments[1] != "counter" && segments[1] != "gauge" {
		http.Error(w, "Тип метрики может быть только counter или gauge, пожалуйста, попробуйте ещё раз", http.StatusBadRequest)
		return "", "", "", fmt.Errorf("некорректный тип метрики: %s", segments[1])
	}
	if segments[2] == "" {
		http.Error(w, "Имя не должно быть пустым, пожалуйста, попробуйте ещё раз", http.StatusNotFound)
		return "", "", "", fmt.Errorf("некорректное имя метрики: %s", segments[2])
	}
	if segments[1] == "counter" {
		if _, err := strconv.ParseInt(segments[3], 10, 64); err != nil {
			http.Error(w, "Значение должно быть целым числом, пожалуйста, попробуйте ещё раз", http.StatusBadRequest)
			return "", "", "", err
		}
	} else if segments[1] == "gauge" {
		if _, err := strconv.ParseFloat(segments[3], 64); err != nil {
			http.Error(w, "Значение должно быть числом с точкой, пожалуйста, попробуйте ещё раз", http.StatusBadRequest)
			return "", "", "", err
		}
	}
	return segments[1], segments[2], segments[3], nil
}
