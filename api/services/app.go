package services

import (
	"fmt"

	"git.houseofkummer.com/lior/home-dns/api/services/grpc"
	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/records"
	"git.houseofkummer.com/lior/home-dns/api/services/resolver"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"git.houseofkummer.com/lior/home-dns/api/services/telemetry"
	"go.uber.org/fx"
)

type Options struct {
	GRPCPort         uint16
	HTTPPort         uint16
	PostgresDatabase string
	PostgresHost     string
	PostgresPassword string
	PostgresPort     uint16
	PostgresUser     string
}

func NewApp(options Options) *fx.App {
	app := fx.New(
		fx.Supply(
			grpc.ListenerOptions{
				Port: options.GRPCPort,
			},
			rest.Options{
				Port: options.HTTPPort,
			},
			storage.Options{
				ConnectionString: fmt.Sprintf(
					"postgres://%s:%s@%s:%d/%s?sslmode=disable",
					options.PostgresUser,
					options.PostgresPassword,
					options.PostgresHost,
					options.PostgresPort,
					options.PostgresDatabase,
				),
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
			logger.Register,
			telemetry.Register,
			resolver.Register,
			records.Register,
		),
		fx.WithLogger(
			logger.NewFxLogger,
		),
	)
	return app
}
