package filterlist

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", "filterlist")
	if err := setup(c); err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}
}

func TestSetupWithArg(t *testing.T) {
	c := caddy.NewTestController("dns", "filterlist foo bar")
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got no errors")
	}
}
