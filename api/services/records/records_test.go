package records

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func TestCreateRecord(t *testing.T) {
	h := createTestHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/records",
		strings.NewReader(`{"zone": "example.com.", "content": "@ A 127.0.0.1", "comment": "test"}`),
	)
	r.Header.Add("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	if w.Result().StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", w.Result().StatusCode)
	}
	var response RecordResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Expected no error, got %v", err)
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

func createTestHandler() http.Handler {
	var handler *chi.Mux
	_ = fx.New(
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
