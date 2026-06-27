package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	RequestIDHeader string = "X-Hookwave-Request-ID"
	RequestIDKey    string = "request_id"
)

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		r.Header.Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
		w.Header().Set(RequestIDHeader, requestID)
	}
	return http.HandlerFunc(fn)
}
