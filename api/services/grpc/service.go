package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func NewService(lc fx.Lifecycle, lis net.Listener) *grpc.Server {
	s := grpc.NewServer()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := s.Serve(lis); err != nil {
				return err
			}
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
