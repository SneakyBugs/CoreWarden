package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

type Options struct {
	Port uint16
}

func NewService(lc fx.Lifecycle, o Options) *chi.Mux {
	r := chi.NewRouter()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", o.Port),
		Handler: r,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				_ = server.ListenAndServe()
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
	return r
}

func NewMockService(lc fx.Lifecycle) *chi.Mux {
	return chi.NewRouter()
}
