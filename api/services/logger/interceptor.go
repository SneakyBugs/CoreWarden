package logger

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var GRPCLoggerError = errors.New("gRPC logger error")

func LoggerInterceptor(l *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, GRPCLoggerError
		}
		start := time.Now()
		defer func() {
			var requestStatus codes.Code
			if err == nil {
				requestStatus = codes.OK
			} else {
				requestStatus = status.Convert(err).Code()
			}
			l.Info(
				"gRPC request",
				zap.String("clientIP", p.Addr.String()),
				zap.String("method", info.FullMethod),
				zap.String("protocol", "gRPC"),
				zap.Int("status", int(requestStatus)),
				zap.Duration("latency", time.Since(start)),
			)
		}()
		return handler(ctx, req)
	}
}
