package interceptors

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/FollowLille/metrics/internal/logger"
)

// TrustedSubnetInterceptor проверяет, находится ли клиент в доверенной подсети
func TrustedSubnetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	if trustedSubnet == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			logger.Log.Info("trusted subnet is empty")
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			logger.Log.Warn("failed to get client IP address")
			return nil, status.Errorf(codes.Internal, "failed to get client IP address")
		}

		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			logger.Log.Warn("failed to split host and port", zap.String("address", p.Addr.String()), zap.Error(err))
			return nil, status.Errorf(codes.InvalidArgument, "failed to split host and port")
		}

		clientIP := net.ParseIP(host)
		if clientIP == nil {
			logger.Log.Warn("invalid client IP address", zap.String("address", p.Addr.String()))
			return nil, status.Errorf(codes.InvalidArgument, "invalid client IP address")
		}

		_, subnet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Log.Error("invalid trusted subnet", zap.String("subnet", trustedSubnet), zap.Error(err))
			return nil, status.Errorf(codes.Internal, "invalid trusted subnet")
		}

		if !subnet.Contains(clientIP) {
			logger.Log.Warn("client IP is not in trusted subnet", zap.String("ip", clientIP.String()), zap.String("subnet", trustedSubnet))
			return nil, status.Errorf(codes.PermissionDenied, "client IP is not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
