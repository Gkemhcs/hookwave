package middleware

import (
	"context"
	"time"

	"github.com/Gkemhcs/hookwave/server/internal/observability"
)

const (
	loggerKey = "logger"
)

type RequestHttpContext struct {
	ctx context.Context
}

func NewRequestHttpContext(ctx context.Context) *RequestHttpContext {
	return &RequestHttpContext{
		ctx: ctx,
	}

}
func (req *RequestHttpContext) SetValue(key any, value any) {
	req.ctx = context.WithValue(req.ctx, key, value)
}

func (req *RequestHttpContext) Deadline() (deadline time.Time, ok bool) {
	return req.ctx.Deadline()
}

func (req *RequestHttpContext) Done() <-chan struct{} {
	return req.ctx.Done()
}

func (req *RequestHttpContext) Err() error {
	return req.ctx.Err()
}

func (req *RequestHttpContext) Value(key any) any {
	value := req.ctx.Value(key)
	return value
}

func (req *RequestHttpContext) AddLogger(logger *observability.Logger) {
	req.ctx = context.WithValue(req.ctx, loggerKey, logger)
}
func (req *RequestHttpContext) Logger() *observability.Logger {

	logger := req.ctx.Value(loggerKey).(*observability.Logger)
	return logger

}
