package records

import (
	"git.houseofkummer.com/lior/home-dns/api/services/enforcer"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type service struct {
	enforcer enforcer.RequestEnforcer
	handler  RecordsStorage
	logger   *zap.Logger
}

func Register(r *chi.Mux, e enforcer.Enforcer, s storage.Storage, l *zap.Logger) {
	sr := service{
		enforcer: enforcer.NewRequestEnforcer(e, "records"),
		handler:  s,
		logger:   l,
	}
	r.Post("/v1/records", sr.HandleCreate())
	r.Get("/v1/records/{id}", sr.HandleRead())
	r.Put("/v1/records/{id}", sr.HandleUpdate())
}
