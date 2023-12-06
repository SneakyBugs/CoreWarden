package grpc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type Options struct {
	Port uint16
}

func NewService(lc fx.Lifecycle, options Options) *grpc.Server {
	s := grpc.NewServer()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", options.Port))
			if err != nil {
				return err
			}
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
