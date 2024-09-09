package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/FollowLille/metrics/internal/storage"
)

func TestHomeHandler(t *testing.T) {
	s := storage.NewMemStorage()
	s.UpdateGauge("testGauge", 123.45)
	s.UpdateCounter("testCounter", 123)

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		HomeHandler(c, s)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testGauge")
	assert.Contains(t, w.Body.String(), "123.45")
	assert.Contains(t, w.Body.String(), "testCounter")
	assert.Contains(t, w.Body.String(), "123")
}

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValue    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "update_counter_success",
			metricType:     "counter",
			metricName:     "myCounter",
			metricValue:    "10",
			expectedStatus: http.StatusOK,
			expectedBody:   "counter updated",
		},
		{
			name:           "update_gauge_success",
			metricType:     "gauge",
			metricName:     "myGauge",
			metricValue:    "123.45",
			expectedStatus: http.StatusOK,
			expectedBody:   "gauge updated",
		},
		{
			name:           "invalid_metric_type",
			metricType:     "invalid",
			metricName:     "myCounter",
			metricValue:    "10",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "metric type must be counter or gauge",
		},
		{
			name:           "invalid_counter_value",
			metricType:     "counter",
			metricName:     "myCounter",
			metricValue:    "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "metric value must be integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storage.NewMemStorage()

			router := gin.Default()
			router.POST("/update/:type/:name/:value", func(c *gin.Context) {
				UpdateHandler(c, s)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/update/"+tt.metricType+"/"+tt.metricName+"/"+tt.metricValue, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestGetValueHandler(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "get_counter_success",
			metricType:     "counter",
			metricName:     "myCounter",
			expectedStatus: http.StatusOK,
			expectedBody:   "10",
		},
		{
			name:           "get_gauge_success",
			metricType:     "gauge",
			metricName:     "myGauge",
			expectedStatus: http.StatusOK,
			expectedBody:   "123.45",
		},
		{
			name:           "counter_not_found",
			metricType:     "counter",
			metricName:     "unknownCounter",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "counter with name unknownCounter not found",
		},
		{
			name:           "gauge_not_found",
			metricType:     "gauge",
			metricName:     "unknownGauge",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "gauge with name unknownGauge not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := storage.NewMemStorage()
			s.UpdateCounter("myCounter", 10)
			s.UpdateGauge("myGauge", 123.45)

			router := gin.Default()
			router.GET("/value/:type/:name", func(c *gin.Context) {
				GetValueHandler(c, s)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/value/"+tt.metricType+"/"+tt.metricName, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
