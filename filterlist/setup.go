package filterlist

import (
	"fmt"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/go-co-op/gocron"
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
				for _, url := range args {
					listURLs = append(listURLs, url)
				}
			default:
				return plugin.Error("filterlist", c.Errf("unknown property %q", c.Val()))
			}
		}
	}
	if len(listURLs) == 0 {
		return plugin.Error("filterlist", c.Errf("blocklists property is required"))
	}

	// No backoff when fetching in init to crash pod.
	engine, err := CreateEngineFromRemote(listURLs, 5, time.Minute*5, 5)
	if err != nil {
		// TODO Log
		fmt.Printf("init error %v\n", err)
		return err
	}
	filterlistPlugin := FilterList{Engine: engine}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		filterlistPlugin.Next = next
		return filterlistPlugin
	})

	cron := gocron.NewScheduler(time.UTC)
	_, err = cron.Every(6).Hours().Do(func() {
		engine, err := CreateEngineFromRemote(listURLs, 5, time.Minute*5, 15)
		if err != nil {
			// TODO Log
			fmt.Printf("cron error %v\n", err)
			return
		}
		filterlistPlugin.Engine = engine
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
