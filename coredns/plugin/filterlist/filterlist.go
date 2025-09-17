package filterlist

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/AdguardTeam/urlfilter"
	agfilter "github.com/AdguardTeam/urlfilter/filterlist"
	"github.com/AdguardTeam/urlfilter/rules"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const name = "filterlist"
const ttl = 604800

type FilterList struct {
	Next   plugin.Handler
	Engine *urlfilter.DNSEngine
	Logger *zap.Logger
}

func (fl FilterList) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	if fl.Engine == nil {
		return plugin.NextOrFailure(fl.Name(), fl.Next, ctx, w, r)
	}

	state := request.Request{W: w, Req: r}
	hostname := strings.TrimSuffix(state.Name(), ".")

	requestsTotal.Add(1)
	matchResult, ok := fl.Engine.MatchRequest(&urlfilter.DNSRequest{
		Hostname: hostname,
		DNSType:  state.QType(),
	})
	if !ok {
		return plugin.NextOrFailure(fl.Name(), fl.Next, ctx, w, r)
	}
	listID, ok := getMatchingListID(matchResult)
	if ok {
		m := new(dns.Msg)
		m.SetReply(r)
		hdr := dns.RR_Header{Name: state.QName(), Ttl: ttl, Class: dns.ClassINET, Rrtype: dns.TypeA}
		m.Answer = []dns.RR{&dns.A{Hdr: hdr, A: net.ParseIP("0.0.0.0").To4()}}
		m.Rcode = dns.RcodeSuccess
		requestsBlocked.Add(1)
		fl.Logger.Info("request blocked",
			zap.String("name", state.Name()),
			zap.Uint64("blocklist", listID),
		)
		return m.Rcode, w.WriteMsg(m)

	}
	// Only DNS rewrite rules were matched.
	// We ignore them, as they may lead to DNS hijack through blocklists.
	return plugin.NextOrFailure(fl.Name(), fl.Next, ctx, w, r)
}

func getMatchingListID(result *urlfilter.DNSResult) (uint64, bool) {
	if result.NetworkRule != nil {
		return uint64(result.NetworkRule.GetFilterListID()), true
	}
	if result.HostRulesV4 != nil {
		return uint64(result.HostRulesV4[0].GetFilterListID()), true
	}
	if result.HostRulesV6 != nil {
		return uint64(result.HostRulesV6[0].GetFilterListID()), true
	}
	return 0, false
}

func (fl FilterList) Name() string {
	return name
}

func CreateEngine(ruleContents []string) (*urlfilter.DNSEngine, error) {
	ruleLists := []agfilter.Interface{}
	for i, ruleList := range ruleContents {
		ruleLists = append(ruleLists, agfilter.NewString(&agfilter.StringConfig{
			RulesText:      ruleList,
			ID:             rules.ListID(i),
			IgnoreCosmetic: true,
		}))
	}
	ruleStorage, err := agfilter.NewRuleStorage(ruleLists)
	if err != nil {
		return nil, err
	}
	return urlfilter.NewDNSEngine(ruleStorage), nil
}

func CreateEngineFromRemote(
	urls []string,
	failuresUntilBackoff int,
	backoffDuration time.Duration,
	failuresUntilError int,
) (*urlfilter.DNSEngine, error) {
	lists := []string{}
	anyFailed := false
	for _, url := range urls {
		fetcher := URLFetcher{
			url: url,
		}
		retrier := Retrier{
			fetcher: fetcher,
			sleeper: RealSleeper{},
		}
		res, err := retrier.FetchWithRetryAndBackoff(failuresUntilBackoff, backoffDuration, failuresUntilError)
		if err != nil {
			anyFailed = true
			continue
		}
		lists = append(lists, res)
	}
	engine, err := CreateEngine(lists)
	if err != nil {
		return nil, err
	}
	if anyFailed {
		return engine, fmt.Errorf("partially fetched lists, engine may still be used")
	}
	return engine, nil
}
