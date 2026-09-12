// Package middleware provides chi-compatible HTTP middleware for request
// IDs, request-scoped logger injection, and per-request access logging.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header used to read/propagate the request ID.
	RequestIDHeader string = "X-Hookwave-Request-ID"
	// RequestIDKey is the context key the request ID is stored under.
	RequestIDKey string = "request_id"
)

// RequestID ensures every request has an ID: it reuses the caller-supplied
// RequestIDHeader value if present, otherwise generates a new one, then
// echoes it back on the response and stores it in the request context.
func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		r.Header.Set(RequestIDHeader, requestID)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
