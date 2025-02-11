package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sneakybugs/corewarden/api/services/logger"
	"github.com/sneakybugs/corewarden/api/services/rest"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/fx"
)

func TestAuthMiddleware(t *testing.T) {
	h := createTestHandler()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	MockLogin(r, "foo")

	h.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status to be 200, got %d", w.Result().StatusCode)
	}
}

func TestAuthMiddlewareUnauthenticated(t *testing.T) {
	h := createTestHandler()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/me", nil)

	h.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status to be 401, got %d", w.Result().StatusCode)
	}
}

func TestAuthMiddlewareServerError(t *testing.T) {
	h := createTestHandler()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.Header.Set(testErrorHeader, "mock error")

	h.ServeHTTP(w, r)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status to be 500, got %d", w.Result().StatusCode)
	}
}

func createTestHandler() http.Handler {
	var handler *chi.Mux
	_ = fx.New(
		fx.Provide(
			logger.NewService,
			rest.NewMockService,
			NewMockAuthenticator,
			NewService,
		),
		fx.Invoke(
			func(rRoot *chi.Mux, a Service) {
				rRoot.Group(func(r chi.Router) {
					r.Use(a.Middleware())
					r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
						sub, ok := GetSubject(r.Context())
						if !ok {
							render.Status(r, http.StatusTeapot)
							render.PlainText(w, r, sub)
							return
						}
						render.PlainText(w, r, sub)
					})
				})
			},
		),
		fx.Populate(
			&handler,
		),
	)
	return handler
}
