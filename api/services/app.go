package services

import (
	"fmt"

	"git.houseofkummer.com/lior/home-dns/api/services/auth"
	"git.houseofkummer.com/lior/home-dns/api/services/enforcer"
	"git.houseofkummer.com/lior/home-dns/api/services/grpc"
	"git.houseofkummer.com/lior/home-dns/api/services/health"
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
	PolicyFile       string
	ServiceAccounts  []auth.ServiceAccount
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
			enforcer.CasbinEnforcerOptions{
				PolicyFile: options.PolicyFile,
			},
			auth.ServiceAccountAuthenticatorOptions{
				Accounts: options.ServiceAccounts,
			},
		),
		fx.Provide(
			grpc.NewListener,
			grpc.NewService,
			rest.NewService,
			health.NewReadinessChecks,
			logger.NewService,
			storage.NewService,
			enforcer.NewCasbinEnforcer,
			auth.NewServiceAccountAuthenticator,
			auth.NewService,
		),
		fx.Invoke(
			logger.Register,
			telemetry.Register,
			resolver.Register,
			records.Register,
			health.Register,
		),
		fx.WithLogger(
			logger.NewFxLogger,
		),
	)
	return app
}
