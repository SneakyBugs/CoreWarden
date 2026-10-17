package zones

import (
	"github.com/go-chi/chi/v5"
	"github.com/sneakybugs/corewarden/api/services/auth"
	"github.com/sneakybugs/corewarden/api/services/enforcer"
	"go.uber.org/zap"
)

type service struct {
	enforcer enforcer.RequestEnforcer
	zones    []string
	logger   *zap.Logger
}

func Register(r *chi.Mux, e enforcer.Enforcer, o Options, l *zap.Logger, a auth.Service) {
	s := service{
		enforcer: enforcer.NewRequestEnforcer(e, "zones"),
		zones:    Canonical(o.Zones),
		logger:   l,
	}
	r.Group(func(r chi.Router) {
		r.Use(a.Middleware())
		r.Get("/v1/zones", s.HandleList())
	})
}
