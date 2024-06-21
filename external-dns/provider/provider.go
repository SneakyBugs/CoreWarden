package provider

import (
	"context"
	"strings"

	"git.houseofkummer.com/lior/home-dns/client"
	"github.com/miekg/dns"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
	eprovider "sigs.k8s.io/external-dns/provider"
)

type Provider struct {
	eprovider.BaseProvider
	client client.Client
	zones  []string
}

type Configuration struct {
	APIEndpoint string `env:"CLIENT_API_ENDPOINT"`
	ID          string `env:"CLIENT_ID"`
	Secret      string `env:"CLIENT_SECRET"`
	// Comma-separated list of zones to manage records in.
	Zones string `env:"CLIENT_ZONES"`
}

func NewProvider(domainFilter endpoint.DomainFilter, config *Configuration) (eprovider.Provider, error) {
	c := client.NewClient(client.ClientOptions{
		APIEndpoint: config.APIEndpoint,
		ID:          config.ID,
		Secret:      config.Secret,
	})
	return &Provider{client: &c, zones: strings.Split(config.Zones, ",")}, nil
}

func (p *Provider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	records := []client.Record{}
	for _, zone := range p.zones {
		zoneRecords, err := p.client.ListRecords(zone)
		if err != nil {
			return nil, err
		}
		records = append(records, zoneRecords...)
	}
	return groupByNameAndType(records), nil
}

func (p *Provider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	return nil
}

func rrToTarget(rr dns.RR) string {
	if v, ok := rr.(*dns.A); ok {
		return v.A.String()
	}
	if v, ok := rr.(*dns.AAAA); ok {
		return v.AAAA.String()
	}
	if v, ok := rr.(*dns.CNAME); ok {
		return v.Target
	}
	if v, ok := rr.(*dns.SRV); ok {
		return v.Target
	}
	if v, ok := rr.(*dns.TXT); ok {
		return strings.Join(v.Txt, "")
	}
	if v, ok := rr.(*dns.NS); ok {
		return v.Ns
	}
	panic("Only supports A, AAAA, CNAME, SRV, TXT and NS records.")
}

func formatName(zone string, rr dns.RR) string {
	zoneLabels := dns.SplitDomainName(zone)
	nameLabels := dns.SplitDomainName(rr.Header().Name)
	return strings.Join(append(nameLabels, zoneLabels...), ".")
}

func groupByNameAndType(records []client.Record) []*endpoint.Endpoint {
	endpoints := []*endpoint.Endpoint{}
	groups := map[string][]client.Record{}
	for _, record := range records {
		rtype := dns.Type(record.RR.Header().Rrtype).String()
		if !provider.SupportedRecordType(rtype) {
			continue
		}
		groupBy := rtype + record.Zone + record.RR.Header().Name
		if _, ok := groups[groupBy]; !ok {
			groups[groupBy] = []client.Record{}
		}
		groups[groupBy] = append(groups[groupBy], record)
	}

	for _, groupRecords := range groups {
		targets := make([]string, len(groupRecords))
		for i, record := range groupRecords {
			targets[i] = rrToTarget(record.RR)
		}
		rtype := dns.Type(records[0].RR.Header().Rrtype).String()
		endpoints = append(
			endpoints,
			endpoint.NewEndpointWithTTL(
				formatName(groupRecords[0].Zone, groupRecords[0].RR),
				rtype,
				endpoint.TTL(records[0].RR.Header().Ttl),
				targets...))
	}
	return endpoints
}
