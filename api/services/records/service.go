package records

import (
	"github.com/sneakybugs/corewarden/api/services/auth"
	"github.com/sneakybugs/corewarden/api/services/enforcer"
	"github.com/sneakybugs/corewarden/api/services/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type service struct {
	enforcer enforcer.RequestEnforcer
	handler  RecordsStorage
	logger   *zap.Logger
}

func Register(r *chi.Mux, e enforcer.Enforcer, s storage.Storage, l *zap.Logger, a auth.Service) {
	sr := service{
		enforcer: enforcer.NewRequestEnforcer(e, "records"),
		handler:  s,
		logger:   l,
	}
	r.Group(func(r chi.Router) {
		r.Use(a.Middleware())
		r.Get("/v1/records", sr.HandleList())
		r.Post("/v1/records", sr.HandleCreate())
		r.Get("/v1/records/{id}", sr.HandleRead())
		r.Put("/v1/records/{id}", sr.HandleUpdate())
		r.Delete("/v1/records/{id}", sr.HandleDelete())
	})
}
