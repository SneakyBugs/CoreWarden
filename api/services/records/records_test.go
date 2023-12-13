package records

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func TestCreateRecord(t *testing.T) {
	h := createTestHandler(false)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Result().StatusCode)
	}
	var response RecordResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response.Zone != "example.com." {
		t.Errorf("Expected zone to be 'example.com.', got '%s'", response.Zone)
	}
	if response.Content != ".\t3600\tIN\tA\t127.0.0.1" {
		t.Errorf("Expected content to be '@\\tA\\t127.0.0.1', got '%s'", response.Content)
	}
	if response.Comment != "test" {
		t.Errorf("Expected comment to be 'test', got '%s'", response.Comment)
	}
}

func TestCreateRecordMissingZone(t *testing.T) {
	h := createTestHandler(false)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Result().StatusCode)
	}
	var response rest.BadRequestErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response.Fields) != 1 {
		t.Fatalf("Expected fields to of length 1, got %d", len(response.Fields))
	}
	if response.Fields[0].Key != "zone" {
		t.Errorf("Expected fields[0].key to be 'zone', got '%s'", response.Fields[0].Key)
	}
	if response.Fields[0].Message != "required" {
		t.Errorf("Expected fields[0].message to be 'required', got '%s'", response.Fields[0].Message)
	}
}

func TestCreateRecordZoneNotFQDN(t *testing.T) {
	h := createTestHandler(false)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Result().StatusCode)
	}
	var response rest.BadRequestErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response.Fields) != 1 {
		t.Fatalf("Expected fields to of length 1, got %d", len(response.Fields))
	}
	if response.Fields[0].Key != "zone" {
		t.Errorf("Expected fields[0].key to be 'zone', got '%s'", response.Fields[0].Key)
	}
	if response.Fields[0].Message != "must end with '.'" {
		t.Errorf("Expected fields[0].message to be 'required', got '%s'", response.Fields[0].Message)
	}
}

func TestCreateRecordMissingContent(t *testing.T) {
	h := createTestHandler(false)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Result().StatusCode)
	}
	var response rest.BadRequestErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response.Fields) != 1 {
		t.Fatalf("Expected fields to of length 1, got %d", len(response.Fields))
	}
	if response.Fields[0].Key != "content" {
		t.Errorf("Expected fields[0].key to be 'zone', got '%s'", response.Fields[0].Key)
	}
	if response.Fields[0].Message != "required" {
		t.Errorf("Expected fields[0].message to be 'required', got '%s'", response.Fields[0].Message)
	}
}

func TestCreateRecordMalformedContent(t *testing.T) {
	h := createTestHandler(false)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@A127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Result().StatusCode)
	}
	var response rest.BadRequestErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response.Fields) != 1 {
		t.Fatalf("Expected fields to of length 1, got %d", len(response.Fields))
	}
	if response.Fields[0].Key != "content" {
		t.Errorf("Expected fields[0].key to be 'zone', got '%s'", response.Fields[0].Key)
	}
}

func TestCreateRecordServerError(t *testing.T) {
	h := createTestHandler(true)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Result().StatusCode)
	}
	var response rest.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response.Message != rest.InternalServerError.Error() {
		t.Errorf(
			"Expected message to be '%s', got '%s'",
			rest.InternalServerError.Error(),
			response.Message,
		)
	}
}

func createTestHandler(returnErrors bool) http.Handler {
	var handler *chi.Mux
	_ = fx.New(
		fx.Supply(
			storage.MockStorageOptions{
				ReturnErrors: returnErrors,
			},
		),
		fx.Provide(
			rest.NewMockService,
			storage.NewMockService,
		),
		fx.Invoke(
			Register,
		),
		fx.Populate(
			&handler,
		),
	)
	return handler
}
