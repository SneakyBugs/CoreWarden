package injector

import (
	"context"

	"git.houseofkummer.com/lior/home-dns/coredns/plugin/injector/resolver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const name = "injector"

type Injector struct {
	client   resolver.ResolverClient
	upstream Upstream
	logger   *zap.Logger
	next     plugin.Handler
}

type Upstream interface {
	Lookup(ctx context.Context, state request.Request, name string, typ uint16) (*dns.Msg, error)
}

func (i *Injector) Name() string {
	return name
}

func (i *Injector) Ready() bool {
	return true
}

func (i *Injector) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	res, err := i.client.Resolve(ctx, &resolver.Question{
		Name:  state.Name(),
		Qtype: uint32(state.QType()),
	})
	if err != nil {
		if status.Convert(err).Code() == codes.NotFound {
			return plugin.NextOrFailure(i.Name(), i.next, ctx, w, r)
		}

		i.logger.Error("grpc error", zap.Error(err))
		return dns.RcodeServerFailure, err
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeSuccess

	m.Answer, err = parseRRs(res.Answer)
	if err != nil {
		i.logger.Error("RR parsing error", zap.Error(err))
		return dns.RcodeServerFailure, err
	}

	m.Ns, err = parseRRs(res.Ns)
	if err != nil {
		i.logger.Error("RR parsing error", zap.Error(err))
		return dns.RcodeServerFailure, err
	}

	m.Extra, err = parseRRs(res.Extra)
	if err != nil {
		i.logger.Error("RR parsing error", zap.Error(err))
		return dns.RcodeServerFailure, err
	}

	// Perform CNAME lookup if needed.
	if len(m.Answer) == 1 && m.Answer[0].Header().Rrtype == dns.TypeCNAME {
		if record, ok := m.Answer[0].(*dns.CNAME); ok {
			i.logger.Info("Querying upstream for CNAME record", zap.String("target", record.Target), zap.Uint16("qtype", state.QType()))
			if up, err := i.upstream.Lookup(ctx, state, record.Target, state.QType()); err == nil && up != nil {
				m.Truncated = up.Truncated
				m.Answer = append(m.Answer, up.Answer...)
			}
		}
	}

	return m.Rcode, w.WriteMsg(m)
}

func parseRRs(rrs []string) (res []dns.RR, err error) {
	res = make([]dns.RR, len(rrs))
	for i, raw := range rrs {
		res[i], err = dns.NewRR(raw)
		if err != nil {
			return nil, err
		}
	}
	return
}
