package grpc

import (
	"context"
	"fmt"
	"net"

	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewService(lc fx.Lifecycle, lis net.Listener, l *zap.Logger) *grpc.Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(logger.LoggerInterceptor(l)),
	)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := s.Serve(lis); err != nil {
					l.Error(
						"gRPC server error",
						zap.Error(err),
					)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.GracefulStop()
			return nil
		},
	})
	return s
}

type ListenerOptions struct {
	Port uint16
}

func NewListener(lc fx.Lifecycle, options ListenerOptions) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", options.Port))
	return lis, err
}
