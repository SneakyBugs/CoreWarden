package records

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/services/auth"
	"git.houseofkummer.com/lior/home-dns/api/services/enforcer"
	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func TestCreateRecord(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
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
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
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
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
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
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
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
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@A127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
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
	h := createTestHandler(errors.New("mock error"))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
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

func TestCreateRecordUnauthorized(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Result().StatusCode)
	}
}

func TestCreateRecordForbidden(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "bob")
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Result().StatusCode)
	}
}

func TestReadRecord(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Result().StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(
		http.MethodGet,
		"/v1/records/1",
		nil,
	)
	auth.MockLogin(r, "alice")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
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

func TestReadRecordNotFound(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodGet,
		"/v1/records/1",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Result().StatusCode)
	}
}

func TestReadRecordUnauthorized(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Result().StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(
		http.MethodGet,
		"/v1/records/1",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Result().StatusCode)
	}
}

func TestReadRecordForbidden(t *testing.T) {
	h := createTestHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "alice")
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Result().StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest(
		http.MethodGet,
		"/v1/records/1",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	auth.MockLogin(r, "bob")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Result().StatusCode)
	}
}

func createTestHandler(returnError error) http.Handler {
	var handler *chi.Mux
	_ = fx.New(
		fx.Supply(
			storage.MockStorageOptions{
				ReturnError: returnError,
			},
			enforcer.CasbinEnforcerOptions{
				PolicyFile: "test_policy.csv",
			},
		),
		fx.Provide(
			logger.NewService,
			rest.NewMockService,
			storage.NewMockService,
			enforcer.NewCasbinEnforcer,
			auth.NewMockAuthenticator,
		),
		fx.Invoke(
			auth.Register,
			Register,
		),
		fx.Populate(
			&handler,
		),
	)
	return handler
}
