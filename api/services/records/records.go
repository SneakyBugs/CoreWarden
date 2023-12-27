package records

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

type RecordsStorage interface {
	CreateRecord(ctx context.Context, p storage.RecordCreateParameters) (storage.Record, error)
	ReadRecord(ctx context.Context, id int) (storage.Record, error)
	UpdateRecord(ctx context.Context, p storage.RecordUpdateParameters) (storage.Record, error)
	DeleteRecord(ctx context.Context, id int) (storage.Record, error)
	ListRecords(ctx context.Context, zone string) ([]storage.Record, error)
}

func (s service) HandleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := &RecordCreateRequest{}
		if err := render.Bind(r, data); err != nil {
			s.logger.Error("failed to bind body", zap.Error(err))
			rest.RenderError(w, r, err)
			return
		}
		ok, err := s.enforcer.IsAuthorized(r, data.Zone)
		if err != nil {
			s.logger.Error("failed to enforce action", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		if !ok {
			s.logger.Error("unauthorized", zap.Error(err))
			rest.RenderError(w, r, &rest.ForbiddenError)
			return
		}
		record, err := s.handler.CreateRecord(r.Context(), storage.RecordCreateParameters{
			Zone:    data.Zone,
			RR:      data.RR.String(),
			Comment: data.Comment,
		})
		if err != nil {
			s.logger.Error("failed to create record", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		s.logger.Info(
			"record created",
			zap.String("zone", record.Zone),
			zap.String("rr", record.RR),
			zap.String("comment", record.Comment),
		)
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, RecordResponse{
			ID:        record.ID,
			Zone:      record.Zone,
			Content:   record.RR,
			Comment:   record.Comment,
			CreatedAt: record.CreatedAt,
			UpdatedOn: record.ModifiedOn,
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
	} else if !dns.IsFqdn(rc.Zone) {
		fieldErrors = append(fieldErrors, rest.KeyError{
			Key:     "zone",
			Message: "must end with '.'",
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

func (s service) HandleRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordID := chi.URLParam(r, "id")
		if recordID == "" {
			s.logger.Error("empty id URL parameter")
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}
		parsedID, err := strconv.Atoi(recordID)
		if err != nil {
			s.logger.Error("failed parsing ID to int", zap.Error(err))
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}
		rec, err := s.handler.ReadRecord(r.Context(), parsedID)
		if err != nil {
			s.logger.Error("failed to read record", zap.Error(err))
			if errors.Is(err, storage.RecordNotFoundError) {
				rest.RenderError(w, r, &rest.NotFoundError)
				return
			}
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		ok, err := s.enforcer.IsAuthorized(r, rec.Zone)
		if err != nil {
			s.logger.Error("failed to enforce action", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		if !ok {
			s.logger.Error("unauthorized", zap.Error(err))
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, RecordResponse{
			ID:        rec.ID,
			Zone:      rec.Zone,
			Content:   rec.RR,
			Comment:   rec.Comment,
			CreatedAt: rec.CreatedAt,
			UpdatedOn: rec.ModifiedOn,
		})
	}
}

func (s service) HandleUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordID := chi.URLParam(r, "id")
		if recordID == "" {
			s.logger.Error("empty id URL parameter")
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}
		parsedID, err := strconv.Atoi(recordID)
		if err != nil {
			s.logger.Error("failed parsing ID to int", zap.Error(err))
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}
		data := &RecordCreateRequest{}
		if err := render.Bind(r, data); err != nil {
			s.logger.Error("failed to bind body", zap.Error(err))
			rest.RenderError(w, r, err)
			return
		}

		// Enforce authorization on new zone.
		ok, err := s.enforcer.IsAuthorized(r, data.Zone)
		if err != nil {
			s.logger.Error("failed to enforce action", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		if !ok {
			s.logger.Error("unauthorized", zap.Error(err))
			rest.RenderError(w, r, &rest.ForbiddenError)
			return
		}

		// Read old record to enforce authorization.
		existingRecord, err := s.handler.ReadRecord(r.Context(), parsedID)
		if err != nil {
			s.logger.Error("failed to read record", zap.Error(err))
			if errors.Is(err, storage.RecordNotFoundError) {
				rest.RenderError(w, r, &rest.NotFoundError)
				return
			}
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		ok, err = s.enforcer.IsAuthorized(r, existingRecord.Zone)
		if err != nil {
			s.logger.Error("failed to enforce action", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		if !ok {
			s.logger.Error("unauthorized", zap.Error(err))
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}

		rec, err := s.handler.UpdateRecord(r.Context(), storage.RecordUpdateParameters{
			Comment: data.Comment,
			ID:      parsedID,
			RR:      data.RR.String(),
			Zone:    data.Zone,
		})
		if err != nil {
			s.logger.Error("failed updating record", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, RecordResponse{
			ID:        rec.ID,
			Zone:      rec.Zone,
			Content:   rec.RR,
			Comment:   rec.Comment,
			CreatedAt: rec.CreatedAt,
			UpdatedOn: rec.ModifiedOn,
		})
	}
}

func (s service) HandleDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordID := chi.URLParam(r, "id")
		if recordID == "" {
			s.logger.Error("empty id URL parameter")
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}
		parsedID, err := strconv.Atoi(recordID)
		if err != nil {
			s.logger.Error("failed parsing ID to int", zap.Error(err))
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}

		// Read record to enforce authorization.
		existingRecord, err := s.handler.ReadRecord(r.Context(), parsedID)
		if err != nil {
			s.logger.Error("failed to read record", zap.Error(err))
			if errors.Is(err, storage.RecordNotFoundError) {
				rest.RenderError(w, r, &rest.NotFoundError)
				return
			}
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		ok, err := s.enforcer.IsAuthorized(r, existingRecord.Zone)
		if err != nil {
			s.logger.Error("failed to enforce action", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		if !ok {
			s.logger.Error("unauthorized", zap.Error(err))
			rest.RenderError(w, r, &rest.NotFoundError)
			return
		}

		rec, err := s.handler.DeleteRecord(r.Context(), parsedID)
		if err != nil {
			s.logger.Error("failed deleting record", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, RecordResponse{
			ID:        rec.ID,
			Zone:      rec.Zone,
			Content:   rec.RR,
			Comment:   rec.Comment,
			CreatedAt: rec.CreatedAt,
			UpdatedOn: rec.ModifiedOn,
		})
	}

}

func (s service) HandleList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zone := r.URL.Query().Get("zone")
		if zone == "" {
			s.logger.Error("missing zone GET parameter")
			rest.RenderError(w, r, &rest.BadRequestErrorResponse{
				Params: []rest.KeyError{
					{
						Key:     "zone",
						Message: "required",
					},
				},
			})
			return
		}
		if !dns.IsFqdn(zone) {
			s.logger.Error("zone is not FQDN")
			rest.RenderError(w, r, &rest.BadRequestErrorResponse{
				Params: []rest.KeyError{
					{
						Key:     "zone",
						Message: "must be FQDN",
					},
				},
			})
			return
		}

		ok, err := s.enforcer.IsAuthorized(r, zone)
		if err != nil {
			s.logger.Error("failed to enforce action", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		if !ok {
			s.logger.Error("unauthorized", zap.Error(err))
			rest.RenderError(w, r, &rest.ForbiddenError)
			return
		}

		recs, err := s.handler.ListRecords(r.Context(), zone)
		if err != nil {
			s.logger.Error("failed to list records", zap.Error(err))
			rest.RenderError(w, r, &rest.InternalServerError)
			return
		}
		records := make([]RecordResponse, len(recs))
		for i, rec := range recs {
			records[i] = RecordResponse{
				ID:        rec.ID,
				Zone:      rec.Zone,
				Content:   rec.RR,
				Comment:   rec.Comment,
				CreatedAt: rec.CreatedAt,
				UpdatedOn: rec.ModifiedOn,
			}
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, records)
	}
}

type RecordResponse struct {
	ID        int       `json:"id"`
	Zone      string    `json:"zone"`
	Content   string    `json:"content"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedOn time.Time `json:"updatedOn"`
}
