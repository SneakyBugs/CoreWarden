package enforcer

import (
	"errors"
	"net/http"

	"github.com/casbin/casbin"
)

type Action string

const ReadAction Action = "read"
const EditAction Action = "edit"

var ServerError = errors.New("server error")

type Enforcer interface {
	Enforce(sub string, obj string, zone string, act Action) (bool, error)
}

func methodToAction(method string) Action {
	if method == http.MethodGet {
		return ReadAction
	}
	return EditAction
}

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func (a *CasbinEnforcer) Enforce(sub string, obj string, zone string, act Action) (bool, error) {
	return a.enforcer.Enforce(sub, obj, zone, string(act)), nil
}

type CasbinEnforcerOptions struct {
	PolicyFile string
}

func NewCasbinEnforcer(o CasbinEnforcerOptions) Enforcer {
	// TODO embed model file
	enforcer, err := casbin.NewEnforcerSafe("model.conf", o.PolicyFile)
	if err != nil {
		panic(err)
	}
	e := CasbinEnforcer{
		enforcer: enforcer,
	}
	e.enforcer.AddFunction("is_subdomain", SubdomainMatchFunc)
	return &e
}
