package interceptors

import (
	"context"
	"crypto/rsa"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FollowLille/metrics/internal/crypto"
	"github.com/FollowLille/metrics/internal/logger"
)

// CryptoEncodeInterceptor кодирует gRPC-запрос
// Принимает RSA-ключ и возвращает gRPC-запрос с зашифрованными данными
//
// Параметры:
//   - privateKey - RSA-ключ
//
// Возвращаемое значение:
//   - gRPC-запрос с зашифрованными данными
func CryptoDecodeInterceptor(privateKey *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if privateKey == nil {
			logger.Log.Warn("private key is nil")
			return handler(ctx, req)
		}

		decodedReq, err := crypto.DecodeGRPCRequest(req, privateKey)
		if err != nil {
			logger.Log.Error("failed to decode request", zap.String("method", info.FullMethod), zap.Any("request", req), zap.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, "failed to decode request: %v", err)
		}
		return handler(ctx, decodedReq)
	}
}
