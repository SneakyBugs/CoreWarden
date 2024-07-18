package injector

import (
	"git.houseofkummer.com/lior/home-dns/coredns/plugin/injector/resolver"
	"git.houseofkummer.com/lior/home-dns/coredns/plugin/slog"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/upstream"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	plugin.Register(name, setup)
}

func setup(c *caddy.Controller) error {
	target := ""

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "target":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return c.ArgErr()
				}
				target = args[0]
			default:
				return plugin.Error(name, c.Errf("unknown property %q", c.Val()))
			}
		}
	}
	if target == "" {
		return plugin.Error(name, c.Errf("property 'target' is required"))
	}

	logger, ok := slog.LoggerFromController(c)
	if !ok {
		// Use no-op logger when slog plugin is not configured.
		logger = zap.NewNop()
	}

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return plugin.Error(name, err)
	}
	client := resolver.NewResolverClient(conn)

	injectorPlugin := Injector{
		logger:   logger,
		client:   client,
		upstream: upstream.New(),
	}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		injectorPlugin.next = next
		return &injectorPlugin
	})

	// TODO handle CoreDNS server restarts?
	c.OnShutdown(func() error {
		return conn.Close()
	})

	return nil
}
