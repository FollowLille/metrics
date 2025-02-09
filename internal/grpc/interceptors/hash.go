package interceptors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/FollowLille/metrics/internal/logger"
)

// HashInterceptor добавляет хэш к gRPC-запросу

func HashInterceptor(hashKey []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		reqString := fmt.Sprintf("%v", req)

		hash := sha256.Sum256(append([]byte(reqString), hashKey...))
		hashString := hex.EncodeToString(hash[:])

		// Добавляем хэш в контекст для дальнейшего использования
		ctx = context.WithValue(ctx, "request-hash", hashString)

		logger.Log.Info("added hash to gRPC request", zap.String("hash", hashString))

		return handler(ctx, req)
	}
}
