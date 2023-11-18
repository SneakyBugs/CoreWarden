package filterlist

import (
	"context"
	"net"
	"strings"

	"github.com/AdguardTeam/urlfilter"
	agfilter "github.com/AdguardTeam/urlfilter/filterlist"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

const name = "filterlist"
const ttl = 604800

type FilterList struct {
	Next   plugin.Handler
	Engine urlfilter.DNSEngine
}

func (fl FilterList) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	hostname := strings.TrimSuffix(state.Name(), ".")

	matchResult, ok := fl.Engine.MatchRequest(&urlfilter.DNSRequest{
		Hostname: hostname,
		DNSType:  state.QType(),
	})
	if !ok {
		return plugin.NextOrFailure(fl.Name(), fl.Next, ctx, w, r)
	}
	if matchResult.NetworkRule != nil || matchResult.HostRulesV4 != nil || matchResult.HostRulesV6 != nil {
		m := new(dns.Msg)
		m.SetReply(r)
		hdr := dns.RR_Header{Name: state.QName(), Ttl: ttl, Class: dns.ClassINET, Rrtype: dns.TypeA}
		m.Answer = []dns.RR{&dns.A{Hdr: hdr, A: net.ParseIP("0.0.0.0").To4()}}
		m.Rcode = dns.RcodeSuccess
		return m.Rcode, w.WriteMsg(m)

	}
	// Only DNS rewrite rules were matched.
	// We ignore them, as they may lead to DNS hijack through blocklists.
	return plugin.NextOrFailure(fl.Name(), fl.Next, ctx, w, r)
}

func (fl FilterList) Name() string {
	return name
}

func CreateEngine(rules []string) (*urlfilter.DNSEngine, error) {
	ruleLists := []agfilter.RuleList{}
	for i, ruleList := range rules {
		ruleLists = append(ruleLists, &agfilter.StringRuleList{
			RulesText:      ruleList,
			ID:             i,
			IgnoreCosmetic: true,
		})
	}
	ruleStorage, err := agfilter.NewRuleStorage(ruleLists)
	if err != nil {
		return nil, err
	}
	return urlfilter.NewDNSEngine(ruleStorage), nil
}
