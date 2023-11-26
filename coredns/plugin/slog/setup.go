package slog

import (
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"go.uber.org/zap"
)

func init() {
	plugin.Register("slog", setup)
}

func setup(c *caddy.Controller) error {
	c.Next() // clog
	if c.Next() {
		return fmt.Errorf("expected no block or arguments")
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return SLog{
			Next:   next,
			Logger: logger,
		}
	})
	return nil
}
