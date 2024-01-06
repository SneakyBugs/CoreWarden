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
	client resolver.ResolverClient
	logger *zap.Logger
	next   plugin.Handler
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
