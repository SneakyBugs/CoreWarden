package logger

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type Options struct {
	DevelopmentMode bool
}

func NewService(o Options) (*zap.Logger, error) {
	if o.DevelopmentMode {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func NewFxLogger(logger *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{
		Logger: logger,
	}
}

func Register(logger *zap.Logger, r *chi.Mux) {
	r.Use(requestLogger(logger))
}
