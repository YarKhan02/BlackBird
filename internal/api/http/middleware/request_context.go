package middleware

import (
	"net/http"
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ctxKey string

const (
	requestIDKey 	ctxKey = "request_id"
	loggerKey		ctxKey = "logger"
)

func RequestContext(baseLogger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.NewString()
			}

			reqLogger := baseLogger.With(zap.String("request_id", reqID))

			ctx := context.WithValue(r.Context(), requestIDKey, reqID)
			ctx = context.WithValue(ctx, loggerKey, reqLogger)

			w.Header().Set("X-Request-ID", reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}