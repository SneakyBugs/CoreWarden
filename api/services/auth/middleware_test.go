package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/fx"
)

func TestAuthMiddleware(t *testing.T) {
	h := createTestHandler()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.Header.Set(testSubjectHeader, "foo")

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
		),
		fx.Invoke(
			Register,
			func(r *chi.Mux) {
				r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
					sub, ok := GetSubject(r.Context())
					if !ok {
						render.Status(r, http.StatusTeapot)
						render.PlainText(w, r, sub)
						return
					}
					render.PlainText(w, r, sub)
				})
			},
		),
		fx.Populate(
			&handler,
		),
	)
	return handler
}

type MockAuthenticator struct{}

const testSubjectHeader = "X-Test-Sub"
const testErrorHeader = "X-Test-Error"

func (a *MockAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (string, error) {
	msg := r.Header.Get(testErrorHeader)
	if msg != "" {
		return "", ServerError
	}
	sub := r.Header.Get(testSubjectHeader)
	if sub == "" {
		return "", UnauthenticatedError
	}
	return sub, nil
}

func NewMockAuthenticator() Authenticator {
	return &MockAuthenticator{}
}
