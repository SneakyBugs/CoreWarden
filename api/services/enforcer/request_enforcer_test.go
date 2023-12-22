package enforcer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/services/auth"
)

func TestRequestEnforcer(t *testing.T) {
	e := NewMockEnforcer(true, nil)
	rq := NewRequestEnforcer(e, "records")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	contextWithSubject := context.WithValue(r.Context(), auth.SubjectContextKey, "alice")
	r = r.WithContext(contextWithSubject)

	res, err := rq.IsAuthorized(r, "example.com.")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res {
		t.Fatalf("expected alice to be authorized")
	}
}

func TestRequestEnforcerUnauthorized(t *testing.T) {
	e := NewMockEnforcer(false, nil)
	rq := NewRequestEnforcer(e, "records")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	contextWithSubject := context.WithValue(r.Context(), auth.SubjectContextKey, "alice")
	r = r.WithContext(contextWithSubject)

	res, err := rq.IsAuthorized(r, "example.com.")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res {
		t.Fatalf("expected alice to not be authorized")
	}
}

func TestRequestEnforcerError(t *testing.T) {
	e := NewMockEnforcer(true, errors.New("test"))
	rq := NewRequestEnforcer(e, "records")

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	contextWithSubject := context.WithValue(r.Context(), auth.SubjectContextKey, "alice")
	r = r.WithContext(contextWithSubject)

	_, err := rq.IsAuthorized(r, "example.com.")

	if err == nil {
		t.Fatalf("expected an error, got no error")
	}
}

func TestRequestEnforcerNoSubject(t *testing.T) {
	e := NewMockEnforcer(true, errors.New("test"))
	rq := NewRequestEnforcer(e, "records")

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	res, err := rq.IsAuthorized(r, "example.com.")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res {
		t.Fatalf("expected subject to not be authorized")
	}
}

type MockEnforcer struct {
	ok  bool
	err error
}

func (e *MockEnforcer) Enforce(sub string, obj string, zone string, act Action) (bool, error) {
	return e.ok, e.err
}

func NewMockEnforcer(ok bool, err error) Enforcer {
	return &MockEnforcer{
		ok:  ok,
		err: err,
	}
}
