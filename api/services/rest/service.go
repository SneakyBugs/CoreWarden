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
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				err := http.ListenAndServe(fmt.Sprintf(":%d", o.Port), r)
				if err != nil {
					// TODO structured log
					panic(err)
				}
			}()
			return nil
		},
	})
	return r
}
