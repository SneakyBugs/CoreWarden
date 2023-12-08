package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
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
