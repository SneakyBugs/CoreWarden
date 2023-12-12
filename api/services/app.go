package services

import (
	"git.houseofkummer.com/lior/home-dns/api/services/grpc"
	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/records"
	"git.houseofkummer.com/lior/home-dns/api/services/resolver"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"go.uber.org/fx"
)

type Options struct {
}

func NewApp(options Options) *fx.App {
	app := fx.New(
		fx.Supply(
			grpc.ListenerOptions{
				Port: 6969,
			},
			rest.Options{
				Port: 6970,
			},
			storage.Options{
				ConnectionString: "postgres://development:development@localhost:5432/development?sslmode=disable",
			},
		),
		fx.Provide(
			grpc.NewListener,
			grpc.NewService,
			rest.NewService,
			logger.NewService,
			storage.NewService,
		),
		fx.Invoke(
			resolver.Register,
			records.Register,
		),
		// fx.WithLogger(
		// 	logger.NewFxLogger,
		// ),
	)
	return app
}
