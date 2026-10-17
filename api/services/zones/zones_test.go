package zones

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/sneakybugs/corewarden/api/services/auth"
	"github.com/sneakybugs/corewarden/api/services/enforcer"
	"github.com/sneakybugs/corewarden/api/services/logger"
	"github.com/sneakybugs/corewarden/api/services/rest"
	"go.uber.org/fx"
)

func TestListZones(t *testing.T) {
	h := createTestHandler([]string{"example.com.", "example.net."})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/zones", nil)
	auth.MockLogin(r, "alice")
	h.ServeHTTP(w, r)
	validateResponseBody(t, r, w.Result())
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
	var response ZonesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !slices.Equal(response.Zones, []string{"example.com."}) {
		t.Errorf("Expected zones [example.com.], got %v", response.Zones)
	}
}

func TestListZonesNone(t *testing.T) {
	h := createTestHandler([]string{"example.net."})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/zones", nil)
	auth.MockLogin(r, "alice")
	h.ServeHTTP(w, r)
	validateResponseBody(t, r, w.Result())
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
	if body := strings.TrimSpace(w.Body.String()); body != `{"zones":[]}` {
		t.Errorf(`Expected body {"zones":[]}, got %s`, body)
	}
}

func TestListZonesRequiresZonesPolicy(t *testing.T) {
	h := createTestHandler([]string{"example.com.", "example.net."})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/zones", nil)
	auth.MockLogin(r, "bob")
	h.ServeHTTP(w, r)
	validateResponseBody(t, r, w.Result())
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
	if body := strings.TrimSpace(w.Body.String()); body != `{"zones":[]}` {
		t.Errorf(`Expected body {"zones":[]}, got %s`, body)
	}
}

func TestListZonesWithoutRecordsPolicy(t *testing.T) {
	h := createTestHandler([]string{"example.com.", "example.net."})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/zones", nil)
	auth.MockLogin(r, "carol")
	h.ServeHTTP(w, r)
	validateResponseBody(t, r, w.Result())
	var response ZonesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !slices.Equal(response.Zones, []string{"example.net."}) {
		t.Errorf("Expected zones [example.net.], got %v", response.Zones)
	}
}

func TestListZonesUnauthorized(t *testing.T) {
	h := createTestHandler([]string{"example.com."})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/zones", nil)
	h.ServeHTTP(w, r)
	validateResponseBody(t, r, w.Result())
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Result().StatusCode)
	}
}

func TestListZonesNormalized(t *testing.T) {
	h := createTestHandler([]string{"Example.COM"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/zones", nil)
	auth.MockLogin(r, "alice")
	h.ServeHTTP(w, r)
	validateResponseBody(t, r, w.Result())
	var response ZonesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !slices.Equal(response.Zones, []string{"example.com."}) {
		t.Errorf("Expected zones [example.com.], got %v", response.Zones)
	}
}

func createTestHandler(zones []string) http.Handler {
	var handler *chi.Mux
	_ = fx.New(
		fx.Supply(
			Options{Zones: zones},
			enforcer.CasbinEnforcerOptions{
				PolicyFile: "test_policy.csv",
			},
			logger.Options{
				DevelopmentMode: true,
			},
		),
		fx.Provide(
			logger.NewService,
			rest.NewMockService,
			enforcer.NewCasbinEnforcer,
			auth.NewMockAuthenticator,
			auth.NewService,
		),
		fx.Invoke(
			Register,
		),
		fx.Populate(
			&handler,
		),
		fx.NopLogger,
	)
	return handler
}

var docValidator validator.Validator

func validateResponseBody(t *testing.T, r *http.Request, w *http.Response) {
	if docValidator == nil {
		apiSpec, err := os.ReadFile("../../openapi.yaml")
		if err != nil {
			t.Fatalf("Failed reading OpenAPI spec: %v\n", err)
		}
		document, err := libopenapi.NewDocument(apiSpec)
		if err != nil {
			t.Fatalf("Failed to parse OpenAPI spec: %v\n", err)
		}
		validator, validationErrs := validator.NewValidator(document)
		if validationErrs != nil {
			t.Fatalf("Failed to create validator: %v\n", validationErrs)
		}
		docValidator = validator
	}
	valid, errs := docValidator.ValidateHttpResponse(r, w)
	if !valid {
		t.Fatalf("Request body failed OpenAPI spec validation: %v", errs)
	}
}
