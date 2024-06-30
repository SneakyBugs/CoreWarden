package enforcer

import (
	"testing"
)

func TestCasbinEnforcer(t *testing.T) {
	e := NewCasbinEnforcer(CasbinEnforcerOptions{
		PolicyFile: "test_policy.csv",
	})
	authorized, err := e.Enforce("bob", "records", "example.com.", EditAction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !authorized {
		t.Fatalf("expected bob to be authorized")
	}
}

func TestCasbinEnforcerUnauthorized(t *testing.T) {
	e := NewCasbinEnforcer(CasbinEnforcerOptions{
		PolicyFile: "test_policy.csv",
	})
	authorized, err := e.Enforce("bob", "records", "example.net.", EditAction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if authorized {
		t.Fatalf("expected bob to not be authorized")
	}
}

func TestCasbinEnforcerSubdomain(t *testing.T) {
	e := NewCasbinEnforcer(CasbinEnforcerOptions{
		PolicyFile: "test_policy.csv",
	})
	authorized, err := e.Enforce("bob", "records", "bob.example.com.", EditAction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !authorized {
		t.Fatalf("expected bob to be authorized")
	}
}

func TestCasbinEnforcerGroup(t *testing.T) {
	e := NewCasbinEnforcer(CasbinEnforcerOptions{
		PolicyFile: "test_policy.csv",
	})
	authorized, err := e.Enforce("alice", "records", "example.com.", ReadAction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !authorized {
		t.Fatalf("expected bob to be authorized")
	}
}
