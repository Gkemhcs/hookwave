package middleware

import (
	"net/http"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/observability"
)

func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context().(*RequestHttpContext)
		logger := ctx.Logger()
		requestID := r.Header.Get(RequestIDHeader)
		method := r.Method
		logger.Info("received request", observability.NewTag("request_id", requestID),
			observability.NewTag("http_method", method))
		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		logger.Info("completed request",
			observability.NewTag("request_id", requestID),
			observability.NewTag("duration_ms", time.Since(start).Milliseconds()),
		)
	}
	return http.HandlerFunc(fn)
}
