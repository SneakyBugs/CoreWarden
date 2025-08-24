package services

import (
	"fmt"

	"github.com/sneakybugs/corewarden/api/services/auth"
	"github.com/sneakybugs/corewarden/api/services/enforcer"
	"github.com/sneakybugs/corewarden/api/services/grpc"
	"github.com/sneakybugs/corewarden/api/services/health"
	"github.com/sneakybugs/corewarden/api/services/logger"
	"github.com/sneakybugs/corewarden/api/services/records"
	"github.com/sneakybugs/corewarden/api/services/resolver"
	"github.com/sneakybugs/corewarden/api/services/rest"
	"github.com/sneakybugs/corewarden/api/services/storage"
	"github.com/sneakybugs/corewarden/api/services/telemetry"
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
	Verbose          bool
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
			logger.Options{
				DevelopmentMode: options.Verbose,
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
