package zones

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/miekg/dns"
	"github.com/sneakybugs/corewarden/api/services/rest"
	"go.uber.org/zap"
)

type Options struct {
	Zones []string
}

type ZonesResponse struct {
	Zones []string `json:"zones"`
}

func Canonical(zones []string) []string {
	result := make([]string, len(zones))
	for i, zone := range zones {
		result[i] = dns.CanonicalName(zone)
	}
	return result
}

func (s service) HandleList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zones := []string{}
		for _, zone := range s.zones {
			ok, err := s.enforcer.IsAuthorized(r, zone)
			if err != nil {
				s.logger.Error("failed to enforce action", zap.Error(err))
				rest.RenderError(w, r, &rest.InternalServerError)
				return
			}
			if ok {
				zones = append(zones, zone)
			}
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, ZonesResponse{Zones: zones})
	}
}
