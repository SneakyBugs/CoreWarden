package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"git.houseofkummer.com/lior/home-dns/api/services/records"
	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
)

func TestCreateRecord(t *testing.T) {
	m := MockHTTPClient{
		Response: createRecordResponse(t, 1, "example.com.", "@ IN A 127.0.0.1", "example"),
		Error:    nil,
	}
	c := Client{
		httpClient: &m,
		endpoint:   "https://localhost:3080/v1",
		credentials: Credentials{
			ClientID:     "example",
			ClientSecret: "secret",
		},
	}
	r, err := c.CreateRecord(CreateRecordParams{
		Zone:    "example.com.",
		RR:      "@ IN A 127.0.0.1",
		Comment: "example",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if r.Zone != "example.com." {
		t.Errorf("Expected zone to be 'example.com.', got '%s'\n", r.Zone)
	}
	if r.RR != "@ IN A 127.0.0.1" {
		t.Errorf("Expected RR to be '@ IN A 127.0.0.1', got '%s'\n", r.RR)
	}
	if r.Comment != "example" {
		t.Errorf("Expected comment to be 'example', got '%s'\n", r.Comment)
	}
	validateRequest(t, m.LastRequest)
}

func TestCreateRecordUnauthorized(t *testing.T) {
	m := MockAPIErrorHTTPClient{
		Error: &rest.UnauthorizedError,
	}
	c := Client{
		httpClient: &m,
		endpoint:   "https://localhost:3080/v1",
		credentials: Credentials{
			ClientID:     "example",
			ClientSecret: "secret",
		},
	}
	_, err := c.CreateRecord(CreateRecordParams{
		Zone:    "example.com.",
		RR:      "@ IN A 127.0.0.1",
		Comment: "example",
	})
	if err == nil {
		t.Fatalf("Expected an error, got nil\n")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected err to be an APIError\n")
	}
	if apiErr.status != http.StatusUnauthorized {
		t.Fatalf("Expected status to be %d, got %d\n", http.StatusUnauthorized, apiErr.status)
	}
	if apiErr.message != "unauthorized" {
		t.Fatalf("Expected message to be 'unauthorized', got %s\n", apiErr.message)
	}
	validateRequest(t, m.LastRequest)
}

func TestReadRecord(t *testing.T) {
	m := MockHTTPClient{
		Response: createRecordResponse(t, 1, "example.com.", "@ IN A 127.0.0.1", "example"),
		Error:    nil,
	}
	c := Client{
		httpClient: &m,
		endpoint:   "https://localhost:3080/v1",
		credentials: Credentials{
			ClientID:     "example",
			ClientSecret: "secret",
		},
	}
	r, err := c.ReadRecord(1)
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if r.Zone != "example.com." {
		t.Errorf("Expected zone to be 'example.com.', got '%s'\n", r.Zone)
	}
	if r.RR != "@ IN A 127.0.0.1" {
		t.Errorf("Expected RR to be '@ IN A 127.0.0.1', got '%s'\n", r.RR)
	}
	if r.Comment != "example" {
		t.Errorf("Expected comment to be 'example', got '%s'\n", r.Comment)
	}
	validateRequest(t, m.LastRequest)
}

type MockHTTPClient struct {
	LastRequest *http.Request
	Response    *http.Response
	Error       error
}

func (c *MockHTTPClient) Do(r *http.Request) (*http.Response, error) {
	c.LastRequest = r
	return c.Response, c.Error
}

type MockAPIErrorHTTPClient struct {
	LastRequest *http.Request
	Error       error
}

func (c *MockAPIErrorHTTPClient) Do(r *http.Request) (*http.Response, error) {
	c.LastRequest = r
	w := httptest.NewRecorder()
	rest.RenderError(w, r, c.Error)
	return w.Result(), nil
}

func createRecordResponse(t *testing.T, id int, zone string, content string, comment string) *http.Response {
	w := httptest.NewRecorder()
	now := time.Now()
	err := json.NewEncoder(w).Encode(records.RecordResponse{
		ID:        id,
		Zone:      zone,
		Content:   content,
		Comment:   comment,
		CreatedAt: now,
		UpdatedOn: now,
	})
	if err != nil {
		t.Fatalf("error encoding record response: %v\n", err)
	}
	return w.Result()
}

var docValidator validator.Validator

func validateRequest(t *testing.T, r *http.Request) {
	if docValidator == nil {
		apiSpec, err := os.ReadFile("../api/openapi.yaml")
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
	valid, errs := docValidator.ValidateHttpRequest(r)
	if !valid {
		t.Fatalf("Request failed OpenAPI spec validation: %v", errs)
	}
}
