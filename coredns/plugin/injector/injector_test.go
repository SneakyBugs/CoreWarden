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
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{"example.com. IN A 127.0.0.1"},
				Ns:     []string{},
				Extra:  []string{},
			},
			Err: nil,
		},
	})
	i := Injector{
		client: &r,
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
	r.AssertDone()
}

func TestForwardWhenNotFound(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{},
			Err: status.Error(
				codes.NotFound,
				"record not found",
			),
		},
	})
	i := Injector{
		client: &r,
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
	r.AssertDone()
}

func TestUnknownError(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{},
			Err: status.Error(
				codes.Unknown,
				"some error",
			),
		},
	})
	i := Injector{
		client: &r,
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
	r.AssertDone()
}

func TestInvalidAnswerRR(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{"example.com. IN A 127.0.0"},
				Ns:     []string{},
				Extra:  []string{},
			},
			Err: nil,
		},
	})
	i := Injector{
		client: &r,
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
	r.AssertDone()
}

func TestInvalidNsRR(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{"example.com. IN A 127.0.0"},
				Extra:  []string{},
			},
			Err: nil,
		},
	})
	i := Injector{
		client: &r,
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
	r.AssertDone()
}

func TestInvalidExtraRR(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{},
				Extra:  []string{"example.com. IN A 127.0.0"},
			},
			Err: nil,
		},
	})
	i := Injector{
		client: &r,
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
	r.AssertDone()
}

type MockResolver struct {
	resolver.ResolverServer
	actions      []MockResolverAction
	t            *testing.T
	currentIndex int
}

func NewMockResolver(t *testing.T, actions []MockResolverAction) MockResolver {
	return MockResolver{
		actions:      actions,
		currentIndex: 0,
		t:            t,
	}
}

type MockResolverAction struct {
	In     *resolver.Question
	Result *resolver.Response
	Err    error
}

func (r *MockResolver) Resolve(ctx context.Context, in *resolver.Question, opts ...grpc.CallOption) (*resolver.Response, error) {
	if len(r.actions) <= r.currentIndex {
		r.t.Fatalf("Client called Resolve when no more method calls were expected\n")
	}
	current := r.currentIndex
	currentIn := r.actions[current].In
	if currentIn.Name != in.Name {
		r.t.Fatalf("Expected in.Name to be '%s', got '%s'\n", currentIn.Name, in.Name)
	}
	if currentIn.Qtype != in.Qtype {
		r.t.Fatalf("Expected in.Qtype to be %d, got %d\n", currentIn.Qtype, in.Qtype)
	}
	r.currentIndex++
	return r.actions[current].Result, r.actions[current].Err
}

func (r *MockResolver) AssertDone() {
	if r.currentIndex != len(r.actions) {
		r.t.Fatalf("Expected client to call all mock actions, called %d out of %d method calls\n", r.currentIndex, len(r.actions))
	}
}
