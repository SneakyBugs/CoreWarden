package filterlist

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestSetup(t *testing.T) {
	c := caddy.NewTestController("dns", `filterlist {
		blocklists https://example.com
	}`)
	if err := setup(c); err != nil {
		t.Fatalf("expected no errors, got: %v", err)
	}
}

func TestSetupMultipleBlocklists(t *testing.T) {
	c := caddy.NewTestController("dns", `filterlist {
		blocklists https://example.com https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
	}`)
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

func TestSetupNoBlock(t *testing.T) {
	c := caddy.NewTestController("dns", "filterlist")
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got no errors")
	}
}

func TestSetupNoBlocklistsArgs(t *testing.T) {
	c := caddy.NewTestController("dns", `filterlist {
		blocklists
	}`)
	if err := setup(c); err == nil {
		t.Fatalf("expected an error, got no errors")
	}
}
