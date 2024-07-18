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
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	h.AssertDone()
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
	req := new(dns.Msg)
	req.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	h := NewMockHandler(t, []MockHandlerAction{
		{
			In:    *req,
			Out:   *new(dns.Msg),
			Rcode: 2,
			Err:   nil,
		},
	})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
	}

	rec := dnstest.NewRecorder(&test.ResponseWriter{})
	code, err := i.ServeDNS(context.Background(), rec, req)
	if code != dns.RcodeServerFailure {
		t.Errorf("Expected rcode %d, got %d\n", dns.RcodeServerFailure, code)
	}
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	r.AssertDone()
	h.AssertDone()
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
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	h.AssertDone()
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
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	h.AssertDone()
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
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	h.AssertDone()
}

func TestCnameTargetingInjector(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{"example.com. IN CNAME example.net."},
				Ns:     []string{},
				Extra:  []string{},
			},
			Err: nil,
		},
		{
			In: &resolver.Question{
				Name:  "example.net.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{"example.net. IN A 127.0.0.1"},
				Ns:     []string{},
				Extra:  []string{},
			},
			Err: nil,
		},
	})
	h := NewMockHandler(t, []MockHandlerAction{})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	if len(rec.Msg.Answer) != 2 {
		t.Fatalf("Expected answer length to be 2, got %d\n", len(rec.Msg.Answer))
	}
	expectedAnswer, err := dns.NewRR("example.com. IN CNAME example.net.")
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if rec.Msg.Answer[0].String() != expectedAnswer.String() {
		t.Fatalf("Expected answer to be '%s', got '%s'\n", expectedAnswer, rec.Msg.Answer[0])
	}
	expectedAnswer, err = dns.NewRR("example.net. IN A 127.0.0.1")
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if rec.Msg.Answer[1].String() != expectedAnswer.String() {
		t.Fatalf("Expected answer to be '%s', got '%s'\n", expectedAnswer, rec.Msg.Answer[1])
	}
	if len(rec.Msg.Ns) != 0 {
		t.Errorf("Expected ns length to be 0, got %d\n", len(rec.Msg.Ns))
	}
	if len(rec.Msg.Extra) != 0 {
		t.Errorf("Expected extra length to be 0, got %d\n", len(rec.Msg.Extra))
	}
	r.AssertDone()
	h.AssertDone()
}

func TestCnameTargetingUpstream(t *testing.T) {
	r := NewMockResolver(t, []MockResolverAction{
		{
			In: &resolver.Question{
				Name:  "example.com.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{"example.com. IN CNAME example.net."},
				Ns:     []string{},
				Extra:  []string{},
			},
			Err: nil,
		},
		{
			In: &resolver.Question{
				Name:  "example.net.",
				Qtype: uint32(dns.TypeA),
			},
			Result: &resolver.Response{
				Answer: []string{},
				Ns:     []string{},
				Extra:  []string{},
			},
			Err: nil,
		},
	})
	nextIn := new(dns.Msg)
	nextIn.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	nextOut := new(dns.Msg)
	nextOut.Answer = []dns.RR{
		test.A("example.com IN A 127.0.0.1"),
	}
	h := NewMockHandler(t, []MockHandlerAction{
		{
			In:    *nextIn,
			Out:   *nextOut,
			Rcode: dns.RcodeSuccess,
			Err:   nil,
		},
	})
	i := Injector{
		client: &r,
		logger: zap.NewNop(),
		next:   &h,
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
	if len(rec.Msg.Answer) != 2 {
		t.Fatalf("Expected answer length to be 2, got %d\n", len(rec.Msg.Answer))
	}
	expectedAnswer, err := dns.NewRR("example.com. IN CNAME example.net.")
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if rec.Msg.Answer[0].String() != expectedAnswer.String() {
		t.Fatalf("Expected answer to be '%s', got '%s'\n", expectedAnswer, rec.Msg.Answer[0])
	}
	expectedAnswer, err = dns.NewRR("example.net. IN A 127.0.0.1")
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if rec.Msg.Answer[1].String() != expectedAnswer.String() {
		t.Fatalf("Expected answer to be '%s', got '%s'\n", expectedAnswer, rec.Msg.Answer[1])
	}
	if len(rec.Msg.Ns) != 0 {
		t.Errorf("Expected ns length to be 0, got %d\n", len(rec.Msg.Ns))
	}
	if len(rec.Msg.Extra) != 0 {
		t.Errorf("Expected extra length to be 0, got %d\n", len(rec.Msg.Extra))
	}
	r.AssertDone()
	h.AssertDone()
}

type MockHandlerAction struct {
	In    dns.Msg
	Out   dns.Msg
	Rcode int
	Err   error
}

type MockHandler struct {
	actions      []MockHandlerAction
	currentIndex int
	t            *testing.T
}

func (h *MockHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, m *dns.Msg) (int, error) {
	if len(h.actions) <= h.currentIndex {
		h.t.Fatalf("Client called Resolve when no more method calls were expected\n")
	}
	current := h.actions[h.currentIndex]

	if len(m.Question) != len(current.In.Question) {
		h.t.Fatalf("Expected question section length to be %d, got %d", len(current.In.Question), len(m.Question))
	}
	for i, expected := range current.In.Question {
		result := m.Question[i]
		if expected.Name != result.Name {
			h.t.Fatalf("Expected question %d Name to be '%s', got '%s'\n", i, expected.Name, result.Name)
		}
		if expected.Qtype != result.Qtype {
			h.t.Fatalf("Expected question %d Qtype to be %d, got %d\n", i, expected.Qtype, result.Qtype)
		}
		if expected.Qclass != result.Qclass {
			h.t.Fatalf("Expected question %d Qclass to be %d, got %d\n", i, expected.Qclass, result.Qclass)
		}
	}

	err := w.WriteMsg(&current.Out)
	if err != nil {
		h.t.Fatalf("Expected dns.Msg to write without error, got %v\n", err)
	}
	h.currentIndex++
	return current.Rcode, current.Err

}

func (h *MockHandler) Name() string {
	return "mock"
}

func (h *MockHandler) AssertDone() {
	if h.currentIndex != len(h.actions) {
		h.t.Fatalf("Expected client to call all mock actions, called %d out of %d method calls\n", h.currentIndex, len(h.actions))
	}
}

func NewMockHandler(t *testing.T, actions []MockHandlerAction) MockHandler {
	return MockHandler{
		actions:      actions,
		currentIndex: 0,
		t:            t,
	}
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
