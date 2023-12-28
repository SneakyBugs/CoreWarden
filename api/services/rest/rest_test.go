package rest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
)

func TestNotFoundResponse(t *testing.T) {
	mux := NewMockService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/records", nil)
	mux.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected status to be 404, got %d", w.Result().StatusCode)
	}
	ValidateResponseBody(t, r, w.Result())
}

func TestMethodNotAllowedResponse(t *testing.T) {
	mux := NewMockService()
	mux.Post("/v1/records", nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/records", nil)
	mux.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status to be 405, got %d", w.Result().StatusCode)
	}
	ValidateResponseBody(t, r, w.Result())
}

var docValidator validator.Validator

func ValidateResponseBody(t *testing.T, r *http.Request, w *http.Response) {
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
