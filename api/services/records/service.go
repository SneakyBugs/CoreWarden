package records

import (
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type service struct {
	handler RecordsStorage
	logger  *zap.Logger
}

func Register(r *chi.Mux, s storage.Storage) {
	logger, _ := zap.NewProduction()
	sr := service{
		handler: s,
		logger:  logger,
	}
	r.Post("/v1/records", sr.HandleCreate())
}
