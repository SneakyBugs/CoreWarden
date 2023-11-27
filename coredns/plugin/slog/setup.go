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
	config := dnsserver.GetConfig(c)
	if config == nil {
		return fmt.Errorf("failed to get DNS server config")
	}
	config.AddPlugin(func(next plugin.Handler) plugin.Handler {
		return SLog{
			Next:   next,
			Logger: logger,
		}
	})

	hosts := config.ListenHosts
	if len(hosts) == 0 && hosts[0] == "" {
		hosts = []string{"*"}
	}
	c.OnStartup(func() error {
		logger.Info("starting up",
			zap.String("zone", config.Zone),
			zap.String("port", config.Port),
			zap.Strings("hosts", hosts),
		)
		return nil
	})
	c.OnShutdown(func() error {
		logger.Info("shutting down",
			zap.String("zone", config.Zone),
			zap.String("port", config.Port),
			zap.Strings("hosts", hosts),
		)
		return nil
	})
	return nil
}
