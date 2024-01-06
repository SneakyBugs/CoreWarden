package injector

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `injector {
		target example.com:6000
	}`)
	if err := setup(c); err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}
}

func TestSetupNoTarget(t *testing.T) {
	c := caddy.NewTestController("dns", `injector {}`)
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestSetupNoBlock(t *testing.T) {
	c := caddy.NewTestController("dns", `injector`)
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got nil")
	}
}
