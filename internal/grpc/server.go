package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"

	"go.uber.org/zap"

	"github.com/FollowLille/metrics/internal/logger"
	"github.com/FollowLille/metrics/internal/metrics"
	"github.com/FollowLille/metrics/internal/storage"
	pb "github.com/FollowLille/metrics/proto"
)

type Server struct {
	pb.UnimplementedMetricsServiceServer
	storage *storage.MemStorage
	mu      sync.Mutex
}

// NewServer инициализирует сервер
func NewServer(storage *storage.MemStorage) *Server {
	return &Server{
		storage: storage,
	}
}

// SendMetrics обрабатывает запрос на отправку метрик
func (s *Server) SendMetrics(ctx context.Context, req *pb.MetricsRequest) (*pb.SendMetricsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request canceled: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var updatedMetrics []*pb.Metric
	var errors []error
	for _, metric := range req.Metrics {
		switch metric.Mtype {
		case metrics.Counter:
			if metric.Delta == nil {
				logger.Log.Warn("delta is nil", zap.String("name", metric.Name))
				errors = append(errors, fmt.Errorf("delta is nil for metric: %s", metric.Name))
				continue
			}
			s.storage.UpdateCounter(metric.Name, *metric.Delta)
		case metrics.Gauge:
			if metric.Value == nil {
				logger.Log.Warn("value is nil", zap.String("name", metric.Name))
				errors = append(errors, fmt.Errorf("value is nil for metric: %s", metric.Name))
				continue
			}
			s.storage.UpdateGauge(metric.Name, *metric.Value)
		default:
			logger.Log.Warn("unknown metric type", zap.String("type", metric.Mtype))
			errors = append(errors, fmt.Errorf("unknown metric type: %s", metric.Mtype))
			continue
		}
		updatedMetrics = append(updatedMetrics, metric)
	}

	if len(errors) > 0 {
		errorMessage := "errors while updating metrics"
		for _, err := range errors {
			errorMessage += "\n" + err.Error()
		}
		return nil, status.Errorf(codes.InvalidArgument, errorMessage)
	}
	return &pb.SendMetricsResponse{Metrics: updatedMetrics}, nil
}

// GetMetrics обрабатывает запрос на получение метрик
func (s *Server) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {

	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request canceled: %v", err)
	}

	var metrics []*pb.Metric
	if req.Filter == "" {
		for name, value := range s.storage.GetAllGauges() {
			metrics = append(metrics, &pb.Metric{
				Name:  name,
				Mtype: "gauge",
				Value: &value,
			})
		}
		for name, value := range s.storage.GetAllCounters() {
			metrics = append(metrics, &pb.Metric{
				Name:  name,
				Mtype: "counter",
				Delta: &value,
			})
		}
	} else {
		gaugeValue, exists := s.storage.GetGauge(req.Filter)
		if exists {
			metrics = append(metrics, &pb.Metric{
				Name:  req.Filter,
				Mtype: "gauge",
				Value: &gaugeValue,
			})
		}

		counterValue, exists := s.storage.GetCounter(req.Filter)
		if exists {
			metrics = append(metrics, &pb.Metric{
				Name:  req.Filter,
				Mtype: "counter",
				Delta: &counterValue,
			})
		}
	}

	return &pb.GetMetricsResponse{
		Metrics: metrics,
	}, nil
}
