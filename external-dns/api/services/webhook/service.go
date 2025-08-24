package webhook

import (
	"io"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/sneakybugs/corewarden/external-dns/provider"
	"go.uber.org/zap"
	"sigs.k8s.io/external-dns/provider/webhook/api"
)

type Options struct {
	APIEndpoint string
	ID          string
	Secret      string
	Zones       string
	BindAddress string
}

func Register(o Options, r *chi.Mux, l *zap.Logger) error {
	p, err := provider.NewProvider(&provider.Configuration{
		APIEndpoint: o.APIEndpoint,
		ID:          o.ID,
		Secret:      o.Secret,
		Zones:       o.Zones,
		Logger:      l,
	})
	if err != nil {
		return err
	}
	ws := api.WebhookServer{
		Provider: p,
	}

	// Replace logger used by ExternalDNS with our structured format.
	logrus.SetOutput(io.Discard)
	logrus.AddHook(NewZapConversionHook(l))

	r.HandleFunc("/", ws.NegotiateHandler)
	r.HandleFunc("/records", ws.RecordsHandler)
	r.HandleFunc("/adjustendpoints", ws.AdjustEndpointsHandler)
	return nil
}
