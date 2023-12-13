package logger

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Options struct {
	DevelopmentMode bool
}

func NewService() (*zap.Logger, error) {
	return zap.NewProduction()
}

func NewFxLogger(logger *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{
		Logger: logger,
	}
}

func Register(logger *zap.Logger, r *chi.Mux, g *grpc.Server) {
	r.Use(requestLogger(logger))
}
