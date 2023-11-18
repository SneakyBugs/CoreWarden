package filterlist

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"
)

func TestFilterlist(t *testing.T) {
	engine, err := CreateEngine([]string{"0.0.0.0 google.com\nfacebook.com\n||example.com^"})
	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}
	fl := FilterList{
		Engine: *engine,
	}
	tests := []struct {
		qname        string
		qtype        uint16
		expectedCode int
		answer       dns.RR
	}{
		{
			qname:        "google.com",
			qtype:        dns.TypeA,
			expectedCode: dns.RcodeSuccess,
			answer:       test.A("google.com. IN A 0.0.0.0"),
		},
		{
			qname:        "facebook.com",
			qtype:        dns.TypeA,
			expectedCode: dns.RcodeSuccess,
			answer:       test.A("facebook.com. IN A 0.0.0.0"),
		},
		{
			qname:        "foo.example.com",
			qtype:        dns.TypeA,
			expectedCode: dns.RcodeSuccess,
			answer:       test.A("foo.example.com. IN A 0.0.0.0"),
		},
	}

	for i, tc := range tests {
		req := new(dns.Msg)
		req.SetQuestion(dns.Fqdn(tc.qname), tc.qtype)
		rec := dnstest.NewRecorder(&test.ResponseWriter{})
		code, err := fl.ServeDNS(context.TODO(), rec, req)
		if err != nil {
			t.Errorf("Test %d: Expected no error, but got %v", i, err)
		}
		if code != tc.expectedCode {
			t.Errorf("Test %d: Expected status code %d, but got %d", i, tc.expectedCode, code)
		}
		if tc.answer == nil && len(rec.Msg.Answer) > 0 {
			t.Errorf("Test %d, expected no answer RR, got %s", i, rec.Msg.Answer[0])
			continue
		}
		if tc.answer != nil {
			if x := tc.answer.Header().Rrtype; x != rec.Msg.Answer[0].Header().Rrtype {
				t.Errorf("Test %d, expected RR type %d in answer, got %d", i, x, rec.Msg.Answer[0].Header().Rrtype)
			}
			if x := tc.answer.Header().Name; x != rec.Msg.Answer[0].Header().Name {
				t.Errorf("Test %d, expected RR name %q in answer, got %q", i, x, rec.Msg.Answer[0].Header().Name)
			}
		}
	}
}
