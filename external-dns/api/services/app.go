package services

import (
	"git.houseofkummer.com/lior/home-dns/api/services/health"
	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/external-dns/api/services/webhook"
	"go.uber.org/fx"
)

type Options struct {
	APIEndpoint string
	ID          string
	Secret      string
	Zones       string
	Port        uint16
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
