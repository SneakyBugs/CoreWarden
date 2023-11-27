package filterlist

import (
	"time"

	"git.houseofkummer.com/lior/home-dns/coredns/plugin/slog"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

func init() {
	plugin.Register("filterlist", setup)
}

func setup(c *caddy.Controller) error {
	// See example config parsing
	// https://github.com/coredns/coredns/blob/master/plugin/transfer/setup.go#L48-L81
	listURLs := []string{}
	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "blocklists":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return c.ArgErr()
				}
				listURLs = append(listURLs, args...)
			default:
				return plugin.Error("filterlist", c.Errf("unknown property %q", c.Val()))
			}
		}
	}
	if len(listURLs) == 0 {
		return plugin.Error("filterlist", c.Errf("blocklists property is required"))
	}

	logger, ok := slog.LoggerFromController(c)
	if !ok {
		// Use no-op logger when slog plugin is not configured.
		logger = zap.NewNop()
	}

	filterlistPlugin := &FilterList{Logger: logger}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		filterlistPlugin.Next = next
		return filterlistPlugin
	})

	cron := gocron.NewScheduler(time.UTC)
	_, err := cron.Every(6).Hours().Do(func() {
		blocklistFetchStart := time.Now()
		engine, err := CreateEngineFromRemote(listURLs, 5, time.Minute*5, 15)
		if err != nil {
			logger.Error("failed to fetch blocklists after retrying 15 times",
				zap.Error(err),
			)
			return
		}
		filterlistPlugin.Engine = engine
		logger.Info("blocklists fetched",
			zap.Duration("duration", time.Since(blocklistFetchStart)),
		)
	})

	c.OnStartup(func() error {
		cron.StartAsync()
		return nil
	})
	c.OnShutdown(func() error {
		cron.Stop()
		return nil
	})
	return err
}
