package auth

import (
	"context"
	"errors"
	"net/http"

	"git.houseofkummer.com/lior/home-dns/api/services/rest"
	"go.uber.org/zap"
)

type contextKey string

const SubjectContextKey contextKey = "sub"

func (s *Service) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sub, err := s.authenticator.Authenticate(w, r)
			if err != nil {
				if errors.Is(err, UnauthenticatedError) {
					s.logger.Error("failed to authenticate", zap.Error(err))
					rest.RenderError(w, r, &rest.UnauthorizedError)
					return
				}
				s.logger.Error("failed to authenticate", zap.Error(err))
				rest.RenderError(w, r, &rest.InternalServerError)
				return
			}

			contextWithSubject := context.WithValue(r.Context(), SubjectContextKey, sub)
			next.ServeHTTP(w, r.WithContext(contextWithSubject))
		})
	}
}

func GetSubject(ctx context.Context) (v string, ok bool) {
	v, ok = ctx.Value(SubjectContextKey).(string)
	return
}
