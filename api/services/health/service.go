package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ReadinessCheck interface {
	Ready() bool
}

type ReadinessChecks struct {
	checks []ReadinessCheck
}

func (rc *ReadinessChecks) Add(c ReadinessCheck) {
	rc.checks = append(rc.checks, c)
}

func NewReadinessChecks() *ReadinessChecks {
	return &ReadinessChecks{
		checks: []ReadinessCheck{},
	}
}

func Register(r *chi.Mux, rc *ReadinessChecks) {
	r.Get("/-/liveness", func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, response{Status: "ok"})
	})
	r.Get("/-/readiness", func(w http.ResponseWriter, r *http.Request) {
		for _, c := range rc.checks {
			if !c.Ready() {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response{Status: "error"})
				return
			}
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, response{Status: "ok"})
	})
}

type response struct {
	Status string `json:"status"`
}
