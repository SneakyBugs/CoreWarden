package injector

import (
	"context"
	"time"

	"github.com/miekg/dns"
	"github.com/sneakybugs/corewarden/coredns/plugin/injector/resolver"
)

type grpcProvider struct {
	client resolver.ResolverClient
}

func (p grpcProvider) Zones(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	res, err := p.client.ListZones(ctx, &resolver.ListZonesRequest{})
	if err != nil {
		return nil, err
	}
	return res.Zones, nil
}

func inZones(zones []string, name string) bool {
	for _, zone := range zones {
		if dns.IsSubDomain(zone, name) {
			return true
		}
	}
	return false
}
