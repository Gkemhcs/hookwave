package middleware

import (
	"net/http"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/api/httpctx"
	"github.com/Gkemhcs/hookwave/server/internal/platform/logging/tag"
)

// Logger logs one "received request" line before and one "completed
// request" line after each request, with its duration. It relies on
// InjectLogger having run earlier in the middleware chain.
func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger, ok := httpctx.LoggerFromContext(r.Context())
		if !ok {
			// No logger attached (e.g. InjectLogger missing earlier in the
			// chain) — don't panic, just skip request logging.
			next.ServeHTTP(w, r)
			return
		}
		requestID := r.Header.Get(RequestIDHeader)
		method := r.Method
		logger.Info("received request", tag.NewTag("request_id", requestID),
			tag.NewTag("http_method", method))
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("completed request",
			tag.NewTag("request_id", requestID),
			tag.NewTag("duration_ms", time.Since(start).Milliseconds()),
		)
	}
	return http.HandlerFunc(fn)
}
