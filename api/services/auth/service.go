package auth

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Service struct {
	authenticator Authenticator
	logger        *zap.Logger
}

func Register(authn Authenticator, l *zap.Logger, h *chi.Mux) {
	s := Service{
		authenticator: authn,
		logger:        l,
	}
	h.Use(s.Middleware())
}
