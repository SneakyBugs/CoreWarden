package slog

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", "slog")
	if err := setup(c); err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}
}

func TestSetupInvalidBlock(t *testing.T) {
	c := caddy.NewTestController("dns", "slog {}")
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got no error")
	}
}

func TestSetupInvalidArgs(t *testing.T) {
	c := caddy.NewTestController("dns", `slog {
		foo bar
	}`)
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got no error")
	}
}
