package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParams(t *testing.T) {
	req, err := paramsToRequest(
		http.MethodGet,
		"https://localhost:3080/v1/records/{{.ID}}",
		UpdateRecordParams{
			ID:      1337,
			Zone:    "example.com",
			RR:      "@ IN A 127.0.0.1",
			Comment: "example",
		},
		Credentials{
			ClientID:     "foo",
			ClientSecret: "foo",
		},
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if req.Method != http.MethodGet {
		t.Fatalf("Expected method to be '%s', got '%s'\n", http.MethodGet, req.Method)
	}
	if req.URL.Path != "/v1/records/1337" {
		t.Fatalf("Expected path to be '/api/v1/records/1337', got '%s'\n", req.URL.Path)
	}
	body := struct {
		Zone    string `json:"zone"`
		RR      string `json:"content"`
		Comment string `json:"comment"`
	}{}
	if req.Body == nil {
		t.Fatalf("Expected req.Body to not be nil")
	}
	readBody, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if err := json.Unmarshal(readBody, &body); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if body.Zone != "example.com" {
		t.Fatalf("Expected body.Zone to be 'example.com', got '%s'\n", body.Zone)
	}
	if body.RR != "@ IN A 127.0.0.1" {
		t.Fatalf("Expected body.RR to be '@ IN A 127.0.0.1', got '%s'\n", body.RR)
	}
	if body.Comment != "example" {
		t.Fatalf("Expected body.Comment to be 'example', got '%s'\n", body.Comment)
	}
	validateRequest(t, req)
}

func TestParseErrorResponseOK(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	apiErr, parsingErr := parseErrorResponse(w.Result(), struct{}{})
	if apiErr != nil {
		t.Fatalf("Expected no error, got %v\n", apiErr)
	}
	if parsingErr != nil {
		t.Fatalf("Expected no error, got %v\n", parsingErr)
	}
}

// TODO Test:
// - error response parsing for 200, 201, ...
// - response parsing error - missing fields in error message? Is it in scope if the server is incompatible?

// func TestParseErrorResponseServerError(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	w.WriteHeader(http.StatusInternalServerError)
// 	w.WriteString(`{"message": "internal server error"}`)
// 	apiErr, parsingErr := parseErrorResponse(w.Result(), struct{}{})
// 	if parsingErr != nil {
// 		t.Fatalf("Expected no error, got %v\n", parsingErr)
// 	}
// 	// TODO Which error?
// }
