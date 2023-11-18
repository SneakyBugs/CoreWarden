package filterlist

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	plugin.Register("filterlist", setup)
}

func setup(c *caddy.Controller) error {
	c.Next() // 'filterlist'
	if c.NextArg() {
		return plugin.Error("filterlist", c.ArgErr())
	}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return FilterList{Next: next}
	})
	return nil
}
