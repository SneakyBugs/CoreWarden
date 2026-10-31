package injector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	_ "github.com/coredns/coredns/plugin/cache"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"github.com/sneakybugs/corewarden/coredns/plugin/injector/resolver"
	"github.com/sneakybugs/corewarden/coredns/plugin/injector/zones"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newCachedChain(t *testing.T, next plugin.Handler) plugin.Handler {
	t.Helper()
	c := caddy.NewTestController("dns", "cache . {\n success 9984 3600\n denial 9984 1800 0\n}")
	setup, err := caddy.DirectiveAction("dns", "cache")
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if err := setup(c); err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	return dnsserver.GetConfig(c).Plugin[0](next)
}

func negativeResponse(rcode int, name string) dns.Msg {
	soa, _ := dns.NewRR("example.com. 3600 IN SOA ns.example.com. noc.example.com. 1 7200 3600 1209600 3600")
	m := dns.Msg{Ns: []dns.RR{soa}}
	m.Response = true
	m.Rcode = rcode
	m.Question = []dns.Question{{Name: name, Qtype: dns.TypeA, Qclass: dns.ClassINET}}
	return m
}

func query(t *testing.T, h plugin.Handler, name string, qtype uint16) *dns.Msg {
	t.Helper()
	req := new(dns.Msg)
	req.SetQuestion(name, qtype)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	if _, err := h.ServeDNS(context.Background(), rec, req); err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	return rec.Msg
}

func TestNegativeResponseIsNotCached(t *testing.T) {
	for name, rcode := range map[string]int{"nodata": dns.RcodeSuccess, "nxdomain": dns.RcodeNameError} {
		t.Run(name, func(t *testing.T) {
			negative := negativeResponse(rcode, "foo.example.com.")
			r := NewMockResolver(t, []MockResolverAction{
				{
					In:  &resolver.Question{Name: "foo.example.com.", Qtype: uint32(dns.TypeA)},
					Err: status.Error(codes.NotFound, "record not found"),
				},
				{
					In:     &resolver.Question{Name: "foo.example.com.", Qtype: uint32(dns.TypeA)},
					Result: &resolver.Response{Answer: []string{"foo.example.com. 60 IN A 10.0.0.1"}},
				},
			})
			h := NewMockHandler(t, []MockHandlerAction{{In: negative, Out: negative, Rcode: rcode}})
			i := Injector{
				zones:  newTestZones("example.com."),
				client: &r,
				logger: zap.NewNop(),
				next:   &h,
			}
			chain := newCachedChain(t, &i)

			first := query(t, chain, "foo.example.com.", dns.TypeA)
			if len(first.Answer) != 0 {
				t.Fatalf("Expected empty answer, got %v\n", first.Answer)
			}
			for _, rr := range first.Ns {
				if rr.Header().Ttl != 0 {
					t.Errorf("Expected authority TTL 0, got %d\n", rr.Header().Ttl)
				}
			}

			second := query(t, chain, "foo.example.com.", dns.TypeA)
			if len(second.Answer) != 1 || second.Answer[0].(*dns.A).A.String() != "10.0.0.1" {
				t.Fatalf("Expected created record, got %v\n", second.Answer)
			}
			r.AssertDone()
			h.AssertDone()
		})
	}
}

func TestPositiveAnswerIsCached(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In:     &resolver.Question{Name: "foo.example.com.", Qtype: uint32(dns.TypeA)},
			Result: &resolver.Response{Answer: []string{"foo.example.com. 60 IN A 10.0.0.1"}},
		},
	})
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		zones:  newTestZones("example.com."),
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
	}
	chain := newCachedChain(t, &i)

	query(t, chain, "foo.example.com.", dns.TypeA)
	second := query(t, chain, "foo.example.com.", dns.TypeA)
	if len(second.Answer) != 1 {
		t.Fatalf("Expected cached answer, got %v\n", second.Answer)
	}
	if ttl := second.Answer[0].Header().Ttl; ttl == 0 || 60 < ttl {
		t.Errorf("Expected TTL in (0, 60], got %d\n", ttl)
	}
	r.AssertDone()
}

func TestNegativeResponseOutsideZonesIsCached(t *testing.T) {
	negative := negativeResponse(dns.RcodeNameError, "foo.example.net.")
	r := NewMockResolver(t, []MockResolverAction{})
	h := NewMockHandler(t, []MockHandlerAction{{In: negative, Out: negative, Rcode: dns.RcodeNameError}})
	i := Injector{
		zones:  newTestZones("example.com."),
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
	}
	chain := newCachedChain(t, &i)

	query(t, chain, "foo.example.net.", dns.TypeA)
	query(t, chain, "foo.example.net.", dns.TypeA)
	h.AssertDone()
}

func TestNegativeResponseIsNotCachedWhenZonesUnavailable(t *testing.T) {
	negative := negativeResponse(dns.RcodeNameError, "foo.example.com.")
	r := NewMockResolver(t, []MockResolverAction{})
	h := NewMockHandler(t, []MockHandlerAction{
		{In: negative, Out: negative, Rcode: dns.RcodeNameError},
		{In: negative, Out: negative, Rcode: dns.RcodeNameError},
	})
	i := Injector{
		zones: zones.NewCache(zones.ProviderFunc(func(context.Context) ([]string, error) {
			return nil, errors.New("api unreachable")
		}), time.Minute),
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
	}
	chain := newCachedChain(t, &i)

	query(t, chain, "foo.example.com.", dns.TypeA)
	query(t, chain, "foo.example.com.", dns.TypeA)
	h.AssertDone()
}

func soaQuery(zone string) dns.Msg {
	m := dns.Msg{}
	m.SetQuestion(zone, dns.TypeSOA)
	return m
}

func TestNoDataFromResolver(t *testing.T) {
	apiSOA := "example.com. 3600 IN SOA ns.example.com. noc.example.com. 1 7200 3600 1209600 3600"
	upstreamSOA, _ := dns.NewRR("example.com. 3600 IN SOA ns.upstream.net. hostmaster.upstream.net. 42 7200 3600 1209600 300")
	upstreamResponse := soaQuery("example.com.")
	upstreamResponse.Response = true
	upstreamResponse.Answer = []dns.RR{upstreamSOA}
	r := NewMockResolver(t, []MockResolverAction{
		{
			In:     &resolver.Question{Name: "foo.example.com.", Qtype: uint32(dns.TypeAAAA)},
			Result: &resolver.Response{Ns: []string{apiSOA}},
		},
		{
			In:     &resolver.Question{Name: "foo.example.com.", Qtype: uint32(dns.TypeAAAA)},
			Result: &resolver.Response{Answer: []string{"foo.example.com. 60 IN AAAA ::1"}},
		},
	})
	h := NewMockHandler(t, []MockHandlerAction{{In: soaQuery("example.com."), Out: upstreamResponse}})
	i := Injector{
		zones:  newTestZones("example.com."),
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
	}
	chain := newCachedChain(t, &i)

	first := query(t, chain, "foo.example.com.", dns.TypeAAAA)
	if first.Rcode != dns.RcodeSuccess {
		t.Fatalf("Expected NOERROR, got %s\n", dns.RcodeToString[first.Rcode])
	}
	if len(first.Answer) != 0 {
		t.Fatalf("Expected empty answer, got %v\n", first.Answer)
	}
	if len(first.Ns) != 1 {
		t.Fatalf("Expected one authority record, got %v\n", first.Ns)
	}
	soa, ok := first.Ns[0].(*dns.SOA)
	if !ok || soa.Ns != "ns.upstream.net." || soa.Serial != 42 {
		t.Fatalf("Expected upstream SOA in authority, got %v\n", first.Ns)
	}
	if soa.Hdr.Ttl != 0 || soa.Minttl != 0 {
		t.Fatalf("Expected SOA TTL and MINIMUM 0, got %d and %d\n", soa.Hdr.Ttl, soa.Minttl)
	}

	second := query(t, chain, "foo.example.com.", dns.TypeAAAA)
	if len(second.Answer) != 1 {
		t.Fatalf("Expected created record, got %v\n", second.Answer)
	}
	r.AssertDone()
	h.AssertDone()
}

func TestNoDataKeepsResolverSOAWhenUpstreamHasNone(t *testing.T) {
	apiSOA := "example.com. 0 IN SOA ns.icann.org. noc.dns.icann.org. 1 7200 3600 1209600 0"
	servfail := soaQuery("example.com.")
	servfail.Response = true
	servfail.Rcode = dns.RcodeServerFailure
	r := NewMockResolver(t, []MockResolverAction{
		{
			In:     &resolver.Question{Name: "foo.example.com.", Qtype: uint32(dns.TypeAAAA)},
			Result: &resolver.Response{Ns: []string{apiSOA}},
		},
	})
	h := NewMockHandler(t, []MockHandlerAction{{In: soaQuery("example.com."), Out: servfail, Rcode: dns.RcodeServerFailure}})
	i := Injector{
		zones:  newTestZones("example.com."),
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
	}

	res := query(t, &i, "foo.example.com.", dns.TypeAAAA)
	if res.Rcode != dns.RcodeSuccess || len(res.Answer) != 0 {
		t.Fatalf("Expected NODATA, got %v\n", res)
	}
	if len(res.Ns) != 1 || res.Ns[0].(*dns.SOA).Ns != "ns.icann.org." {
		t.Fatalf("Expected resolver SOA in authority, got %v\n", res.Ns)
	}
	r.AssertDone()
	h.AssertDone()
}
