package services

import (
	"github.com/sneakybugs/corewarden/api/services/health"
	"github.com/sneakybugs/corewarden/api/services/logger"
	"github.com/sneakybugs/corewarden/api/services/rest"
	"github.com/sneakybugs/corewarden/external-dns/api/services/webhook"
	"go.uber.org/fx"
)

type Options struct {
	APIEndpoint string
	ID          string
	Secret      string
	Zones       string
	Port        uint16
	Verbose     bool
}

func NewApp(options Options) *fx.App {
	app := fx.New(
		fx.Supply(
			rest.Options{
				Port: options.Port,
			},
			webhook.Options{
				APIEndpoint: options.APIEndpoint,
				ID:          options.ID,
				Secret:      options.Secret,
				Zones:       options.Zones,
			},
			logger.Options{
				DevelopmentMode: options.Verbose,
			},
		),
		fx.Provide(
			rest.NewService,
			health.NewReadinessChecks,
			logger.NewService,
		),
		fx.Invoke(
			logger.Register,
			health.Register,
			webhook.Register,
		),
		fx.WithLogger(
			logger.NewFxLogger,
		),
	)
	return app
}
