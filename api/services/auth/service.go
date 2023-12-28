package auth

import (
	"go.uber.org/zap"
)

type Service struct {
	authenticator Authenticator
	logger        *zap.Logger
}

func NewService(authn Authenticator, l *zap.Logger) Service {
	return Service{
		authenticator: authn,
		logger:        l,
	}
}
