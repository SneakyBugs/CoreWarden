package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestServiceAccountAuth(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("test"), 8)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	a := NewServiceAccountAuthenticator(ServiceAccountAuthenticatorOptions{
		Accounts: []ServiceAccount{
			{
				ID:         "foo",
				SecretHash: hash,
			},
		},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("foo", "test")
	sub, err := a.Authenticate(w, r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sub != "foo" {
		t.Fatalf("expected subject to be foo, got %s", sub)
	}
}

func TestServiceAccountAuthBadUsername(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("test"), 8)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	a := NewServiceAccountAuthenticator(ServiceAccountAuthenticatorOptions{
		Accounts: []ServiceAccount{
			{
				ID:         "foo",
				SecretHash: hash,
			},
		},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("bar", "test")
	_, err = a.Authenticate(w, r)
	if err != UnauthenticatedError {
		t.Fatalf("expected %v error, got %v", UnauthenticatedError, err)
	}
}

func TestServiceAccountAuthBadPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("test"), 8)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	a := NewServiceAccountAuthenticator(ServiceAccountAuthenticatorOptions{
		Accounts: []ServiceAccount{
			{
				ID:         "foo",
				SecretHash: hash,
			},
		},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("foo", "wrong")
	_, err = a.Authenticate(w, r)
	if err != UnauthenticatedError {
		t.Fatalf("expected %v error, got %v", UnauthenticatedError, err)
	}
}
