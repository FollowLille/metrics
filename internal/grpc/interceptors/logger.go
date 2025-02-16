package interceptors

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"

	"github.com/FollowLille/metrics/internal/logger"
)

// LoggingInterceptor логирует входящие и исходящие запросы
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	logger.Log.Info("gRPC request", zap.String("method", info.FullMethod), zap.Any("request", req))
	resp, err := handler(ctx, req)

	if err != nil {
		logger.Log.Error("gRPC error", zap.String("method", info.FullMethod), zap.Error(err))
	} else {
		logger.Log.Info("gRPC response", zap.String("method", info.FullMethod), zap.Any("response", resp), zap.Duration("duration", time.Since(start)))
	}

	return resp, err
}
