package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/sneakybugs/corewarden/client"
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

func NewProvider(config *Configuration) (eprovider.Provider, error) {
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

func splitToZoneAndName(domain string, managedZones []string) (string, string, error) {
	for _, zone := range managedZones {
		fmt.Printf("domain: %s, zone: %s\n", domain, zone)
		zoneLabelCount := dns.CountLabel(zone)
		commonLabels := dns.CompareDomainName(domain, zone)
		if commonLabels == zoneLabelCount {
			if dns.CountLabel(domain) == zoneLabelCount {
				return zone, ".", nil
			}
			labels := dns.SplitDomainName(domain)
			fmt.Printf("zone: %s, name: %s, common labels: %d, domain labels: %d\n", zone, strings.Join(labels[:commonLabels], "."), commonLabels, len(labels))
			return zone, strings.Join(labels[:len(labels)-commonLabels], "."), nil
		}
	}
	return "", "", fmt.Errorf("Domains being split must be subdomains of zones in managedZones")
}

func (p *Provider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	errors := []error{}
	for _, endpoint := range changes.Create {
		fmt.Printf("dns name: %s, record type: %s\n", endpoint.DNSName, endpoint.RecordType)
		for _, target := range endpoint.Targets {
			fmt.Printf("dns name: %s, record type: %s, target: %s\n", endpoint.DNSName, endpoint.RecordType, target)
			// Here the error occurs.
			rr, zone, err := endpointToRR(p.zones, *endpoint, target)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			_, err = p.client.CreateRecord(client.CreateRecordParams{
				Zone:    zone,
				RR:      rr.String(),
				Comment: "",
			})
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}
	}
	if len(changes.UpdateNew) == 0 {
		// No more updates to perform, avoid listing records.
		if 0 < len(errors) {
			return fmt.Errorf("Encountered %d recoverable errors", len(errors))
		}
		return nil
	}
	existingRecords := []client.Record{}
	for _, zone := range p.zones {
		zoneRecords, err := p.client.ListRecords(zone)
		if err != nil {
			return err
		}
		existingRecords = append(existingRecords, zoneRecords...)
	}
	for i, desired := range changes.UpdateNew {
		current := changes.UpdateOld[i]
		add, remove, leave := provider.Difference(current.Targets, desired.Targets)
		for _, target := range remove {
			record, err := findRecord(p.zones, existingRecords, *current, target)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			_, err = p.client.DeleteRecord(record.ID)
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}
		for _, target := range add {
			rr, zone, err := endpointToRR(p.zones, *desired, target)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			_, err = p.client.CreateRecord(client.CreateRecordParams{
				Zone:    zone,
				RR:      rr.String(),
				Comment: "",
			})
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}
		for _, target := range leave {
			record, err := findRecord(p.zones, existingRecords, *current, target)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			rr, _, err := endpointToRR(p.zones, *desired, target)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			if record.RR.String() == rr.String() {
				// Record does not need to be updated.
				continue
			}
			_, err = p.client.UpdateRecord(client.UpdateRecordParams{
				ID:      record.ID,
				Zone:    record.Zone,
				RR:      rr.String(),
				Comment: record.Comment,
			})
			if err != nil {
				errors = append(errors, err)
				continue
			}

		}
	}
	if 0 < len(errors) {
		return fmt.Errorf("Encountered %d recoverable errors", len(errors))
	}
	return nil
}

func endpointToRR(zones []string, e endpoint.Endpoint, target string) (dns.RR, string, error) {
	zone, name, err := splitToZoneAndName(dns.Fqdn(e.DNSName), zones)
	if err != nil {
		return nil, "", err
	}
	content := fmt.Sprintf("%s %d IN %s %s", name, e.RecordTTL, e.RecordType, target)
	rr, err := dns.NewRR(content)
	return rr, zone, err
}

func findRecord(zones []string, records []client.Record, e endpoint.Endpoint, target string) (client.Record, error) {
	targetRR, zone, err := endpointToRR(zones, e, target)
	if err != nil {
		return client.Record{}, err
	}
	for _, r := range records {
		if r.RR.String() == targetRR.String() && r.Zone == zone {
			return r, nil
		}
	}
	return client.Record{}, fmt.Errorf("record not found")
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
