package services

import (
	"git.houseofkummer.com/lior/home-dns/api/services/grpc"
	"git.houseofkummer.com/lior/home-dns/api/services/resolver"
	"go.uber.org/fx"
)

type Options struct {
}

func NewApp(options Options) *fx.App {
	app := fx.New(
		fx.Supply(
			grpc.Options{
				Port: 6969,
			},
		),
		fx.Provide(
			grpc.NewService,
		),
		fx.Invoke(
			resolver.Register,
		),
	)
	return app
}
