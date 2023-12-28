package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
)

type Options struct {
	Port uint16
}

func NewService(lc fx.Lifecycle, o Options) *chi.Mux {
	r := newRouter()
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

func NewMockService() *chi.Mux {
	return newRouter()
}

func newRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.NotFound(ErrorHandler(&NotFoundError))
	r.MethodNotAllowed(ErrorHandler(&MethodNotAllowedError))
	return r
}
