package records

import (
	"context"
	"net/http"
	"time"

	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/render"
	"github.com/miekg/dns"
)

type RecordsStorage interface {
	CreateRecord(ctx context.Context, p storage.RecordCreateParameters) (storage.Record, error)
}

func (s service) HandleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &RecordCreateRequest{}
		if err := render.Bind(r, data); err != nil {
			rest.RenderError(w, r, err)
			return
		}
		record, err := s.handler.CreateRecord(r.Context(), storage.RecordCreateParameters{
			Zone:    data.Zone,
			RR:      data.RR.String(),
			Comment: data.Comment,
		})
		if err != nil {
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, RecordResponse{
			ID:        record.ID,
			Zone:      record.Zone,
			Content:   record.RR,
			Comment:   record.Comment,
			CreatedAt: record.CreatedAt,
			UpdatedOn: record.UpdatedOn,
		})
	}
}

type RecordCreateRequest struct {
	Zone    string `json:"zone"`
	Content string `json:"content"`
	Comment string `json:"comment,omitempty"`
	RR      dns.RR `json:"-"`
}

func (rc *RecordCreateRequest) Bind(r *http.Request) error {
	fieldErrors := []rest.KeyError{}
	if rc.Zone == "" {
		fieldErrors = append(fieldErrors, rest.KeyError{
			Key:     "zone",
			Message: "required",
		})
	}
	if rc.Content == "" {
		fieldErrors = append(fieldErrors, rest.KeyError{
			Key:     "content",
			Message: "required",
		})
		// No need to parse content if content is empty.
		return &rest.BadRequestErrorResponse{
			Fields: fieldErrors,
		}
	}
	var err error
	rc.RR, err = dns.NewRR(rc.Content)
	if err != nil {
		fieldErrors = append(fieldErrors, rest.KeyError{
			Key:     "content",
			Message: err.Error(),
		})
	}
	if 0 < len(fieldErrors) {
		return &rest.BadRequestErrorResponse{
			Fields: fieldErrors,
		}
	}
	return nil
}

type RecordResponse struct {
	ID        int       `json:"id"`
	Zone      string    `json:"zone"`
	Content   string    `json:"content"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedOn time.Time `json:"updatedOn"`
}
