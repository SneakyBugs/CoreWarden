package enforcer

import (
	"net/http"

	"git.houseofkummer.com/lior/home-dns/api/services/auth"
	"github.com/miekg/dns"
)

type RequestEnforcer struct {
	enforcer Enforcer
	object   string
}

func (re *RequestEnforcer) IsAuthorized(r *http.Request, zone string) (result bool, err error) {
	sub, ok := auth.GetSubject(r.Context())
	if !ok {
		return false, nil
	}
	result, err = re.enforcer.Enforce(sub, re.object, zone, methodToAction(r.Method))
	return
}

func NewRequestEnforcer(e Enforcer, obj string) RequestEnforcer {
	return RequestEnforcer{
		enforcer: e,
		object:   obj,
	}
}

func SubdomainMatchFunc(args ...any) (any, error) {
	zone := args[0].(string)
	subdomain := args[1].(string)
	return dns.IsSubDomain(zone, subdomain), nil
}
