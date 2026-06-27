package middleware

import (
	"net/http"

	"github.com/Gkemhcs/hookwave/server/internal/observability"
)

func SetRequestHttpContext(logger *observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := NewRequestHttpContext(r.Context())
			reqCtx.AddLogger(logger)

			next.ServeHTTP(w, r.WithContext(reqCtx))
		})
	}
}
