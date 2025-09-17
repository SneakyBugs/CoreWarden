package enforcer

import (
	_ "embed"
	"errors"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

type Action string

const ReadAction Action = "read"
const EditAction Action = "edit"

var ErrServer = errors.New("server error")

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

func (a *CasbinEnforcer) Enforce(sub string, obj string, zone string, act Action) (ok bool, err error) {
	ok, err = a.enforcer.Enforce(sub, obj, zone, string(act))
	return
}

type CasbinEnforcerOptions struct {
	PolicyFile string
}

//go:embed model.conf
var modelConf string

func NewCasbinEnforcer(o CasbinEnforcerOptions) Enforcer {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		panic(err)
	}
	a := fileadapter.NewAdapter(o.PolicyFile)
	enforcer, err := casbin.NewEnforcer(m, a)
	if err != nil {
		panic(err)
	}
	e := CasbinEnforcer{
		enforcer: enforcer,
	}
	e.enforcer.AddFunction("is_subdomain", SubdomainMatchFunc)
	return &e
}
