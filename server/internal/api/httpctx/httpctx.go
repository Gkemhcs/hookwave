// Package httpctx carries request-scoped values (currently: the logger) on the
// standard context.Context, using the normal WithValue/typed-key pattern instead
// of a custom Context implementation.
package httpctx

import (
	"context"

	"github.com/Gkemhcs/hookwave/server/internal/platform/logging"
)

// ctxKey is unexported so no package outside httpctx can set or collide with
// these keys, even accidentally.
type ctxKey int

const loggerKey ctxKey = iota

// WithLogger returns a new context carrying logger. It does not modify ctx.
func WithLogger(ctx context.Context, logger *logging.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext returns the logger previously attached with WithLogger.
// ok is false if no logger was ever attached to ctx.
func LoggerFromContext(ctx context.Context) (logger *logging.Logger, ok bool) {
	logger, ok = ctx.Value(loggerKey).(*logging.Logger)
	return logger, ok
}
