package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func requestLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			defer func() {
				logger.Info(
					"HTTP request",
					zap.String("clientIP", r.RemoteAddr),
					zap.String("method", r.Method),
					zap.String("host", r.Host),
					zap.String("path", r.URL.Path),
					zap.String("protocol", r.Proto),
					zap.Int("status", ww.Status()),
					zap.Duration("latency", time.Since(start)),
					zap.String("userAgent", r.UserAgent()),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
