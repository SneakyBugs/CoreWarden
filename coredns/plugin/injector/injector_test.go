package injector

import (
	"context"
	"strings"
	"testing"

	"git.houseofkummer.com/lior/home-dns/coredns/plugin/injector/resolver"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInjector(t *testing.T) {
	i := Injector{
		client: &mockResolver{
			result: &resolver.Response{
				Answer: []string{"example.com. IN A 127.0.0.1"},
				Ns:     []string{},
				Extra:  []string{},
			},
			err: nil,
		},
		logger: zap.NewNop(),
	}

	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if code != dns.RcodeSuccess {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeSuccess, code)
	}
	if rec.Msg == nil {
		t.Fatalf("Expected message to not be nil\n")
	}
	if len(rec.Msg.Answer) != 1 {
		t.Fatalf("Expected answer length to be 1, got %d\n", len(rec.Msg.Answer))
	}
	expectedAnswer, err := dns.NewRR("example.com. IN A 127.0.0.1")
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if rec.Msg.Answer[0].String() != expectedAnswer.String() {
		t.Fatalf("Expected answer to be '%s', got '%s'\n", expectedAnswer, rec.Msg.Answer[0])
	}
	if len(rec.Msg.Ns) != 0 {
		t.Errorf("Expected ns length to be 0, got %d\n", len(rec.Msg.Ns))
	}
	if len(rec.Msg.Extra) != 0 {
		t.Errorf("Expected extra length to be 0, got %d\n", len(rec.Msg.Extra))
	}
}

func TestForwardWhenNotFound(t *testing.T) {
	i := Injector{
		client: &mockResolver{
			result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{},
				Extra:  []string{},
			},
			err: status.Error(
				codes.NotFound,
				"record not found",
			),
		},
		logger: zap.NewNop(),
	}

	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if code != dns.RcodeServerFailure {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeServerFailure, code)
	}
	if err == nil {
		t.Fatalf("Expected an error\n")
	}
	if !strings.Contains(err.Error(), "no next plugin found") {
		t.Fatalf("Expected error string to contain 'no next plugin found'")
	}
}

func TestUnknownError(t *testing.T) {
	i := Injector{
		client: &mockResolver{
			result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{},
				Extra:  []string{},
			},
			err: status.Error(
				codes.Unknown,
				"some error",
			),
		},
		logger: zap.NewNop(),
	}

	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if code != dns.RcodeServerFailure {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeServerFailure, code)
	}
	if err == nil {
		t.Fatalf("Expected an error\n")
	}
	if !strings.Contains(err.Error(), "some error") {
		t.Fatalf("Expected error string to not contain 'some error'")
	}
}

func TestInvalidAnswerRR(t *testing.T) {
	i := Injector{
		client: &mockResolver{
			result: &resolver.Response{
				Answer: []string{"example.com. IN A 127.0.0"},
				Ns:     []string{},
				Extra:  []string{},
			},
			err: nil,
		},
		logger: zap.NewNop(),
	}

	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if code != dns.RcodeServerFailure {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeServerFailure, code)
	}
	if err == nil {
		t.Fatalf("Expected an error\n")
	}
}

func TestInvalidNsRR(t *testing.T) {
	i := Injector{
		client: &mockResolver{
			result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{"example.com. IN A 127.0.0"},
				Extra:  []string{},
			},
			err: nil,
		},
		logger: zap.NewNop(),
	}

	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if code != dns.RcodeServerFailure {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeServerFailure, code)
	}
	if err == nil {
		t.Fatalf("Expected an error\n")
	}
}

func TestInvalidExtraRR(t *testing.T) {
	i := Injector{
		client: &mockResolver{
			result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{},
				Extra:  []string{"example.com. IN A 127.0.0"},
			},
			err: nil,
		},
		logger: zap.NewNop(),
	}

	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if code != dns.RcodeServerFailure {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeServerFailure, code)
	}
	if err == nil {
		t.Fatalf("Expected an error\n")
	}
}

type mockResolver struct {
	result *resolver.Response
	err    error
	resolver.ResolverServer
}

func (r *mockResolver) Resolve(ctx context.Context, in *resolver.Question, opts ...grpc.CallOption) (*resolver.Response, error) {
	return r.result, r.err
}
