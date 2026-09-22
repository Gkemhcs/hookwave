package middleware

import (
	"net/http"

	"github.com/Gkemhcs/hookwave/server/internal/api/httpctx"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging"
)

// InjectLogger attaches logger to the request context so downstream
// middleware and handlers can retrieve it via httpctx.LoggerFromContext.
func InjectLogger(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := httpctx.WithLogger(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
