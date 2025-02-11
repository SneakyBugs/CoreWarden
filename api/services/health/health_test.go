package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sneakybugs/corewarden/api/services/rest"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func TestLiveness(t *testing.T) {
	h := createTestHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/-/liveness", nil)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d\n", w.Result().StatusCode)
	}
}

func TestReadinessEmpty(t *testing.T) {
	h := createTestHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/-/readiness", nil)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d\n", w.Result().StatusCode)
	}
}

func TestReadinessOK(t *testing.T) {
	h := createTestHandler(
		fx.Invoke(
			registerSuccessfulCheck,
		),
	)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/-/readiness", nil)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d\n", w.Result().StatusCode)
	}
}

func TestReadinessBad(t *testing.T) {
	h := createTestHandler(
		fx.Invoke(
			registerFailingCheck,
		),
	)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/-/readiness", nil)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status code 500, got %d\n", w.Result().StatusCode)
	}
}

func TestReadinessMultipleOK(t *testing.T) {
	h := createTestHandler(
		fx.Invoke(
			registerSuccessfulCheck,
			registerSuccessfulCheck,
			registerSuccessfulCheck,
		),
	)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/-/readiness", nil)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d\n", w.Result().StatusCode)
	}
}

func TestReadinessSingleBad(t *testing.T) {
	h := createTestHandler(
		fx.Invoke(
			registerSuccessfulCheck,
			registerSuccessfulCheck,
			registerFailingCheck,
		),
	)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/-/readiness", nil)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status code 500, got %d\n", w.Result().StatusCode)
	}
}

func createTestHandler(decorators ...fx.Option) http.Handler {
	var handler *chi.Mux
	defaultOptions := []fx.Option{
		fx.Provide(
			rest.NewMockService,
			NewReadinessChecks,
		),
		fx.Invoke(
			Register,
		),
		fx.Populate(
			&handler,
		),
	}
	_ = fx.New(
		append(defaultOptions, decorators...)...,
	)
	return handler
}

type stubSuccessfulCheck struct{}

func (c *stubSuccessfulCheck) Ready() bool {
	return true
}

func registerSuccessfulCheck(rc *ReadinessChecks) {
	rc.Add(&stubSuccessfulCheck{})
}

type stubFailingCheck struct{}

func (c *stubFailingCheck) Ready() bool {
	return false
}
func registerFailingCheck(rc *ReadinessChecks) {
	rc.Add(&stubFailingCheck{})
}
