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
	fmt.Println("Setup is called")
	c.Next() // 'filterlist'
	if c.NextArg() {
		return plugin.Error("filterlist", c.ArgErr())
	}

	listURLs := []string{
		"https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt",
		// "https://adaway.org/hosts.txt",
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
	cron.Every(6).Hours().Do(func() {
		engine, err := CreateEngineFromRemote(listURLs, 5, time.Minute*5, 15)
		if err != nil {
			// TODO Log
			fmt.Printf("cron error %v\n", err)
			return
		}
		filterlistPlugin.Engine = engine
	})
	c.OnShutdown(func() error {
		cron.Stop()
		return nil
	})
	cron.StartAsync()
	return nil
}
