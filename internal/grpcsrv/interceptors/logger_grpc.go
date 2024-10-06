package interceptors

import (
	"context"
	"go.uber.org/zap"
	"time"

	"github.com/SversusN/shortener/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggerInterceptor логирует входящие запросы.
func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	lg := logger.CreateLogger(zap.NewAtomicLevelAt(zap.InfoLevel))
	sl := *lg.Logger.Sugar()
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	status, _ := status.FromError(err)
	sl.Infoln(
		"gRPC request",
		"method", info.FullMethod,
		"duration", duration,
		"code", status.Code(),
	)
	return resp, err
}
