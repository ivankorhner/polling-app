package logging

import (
	"context"
	"log/slog"
)

type ctxKey string

const slogFields ctxKey = "slogFields"

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, attr := range attrs {
			r.AddAttrs(attr)
		}
	}
	return h.Handler.Handle(ctx, r)
}

func AppendCtx(ctx context.Context, attrs ...slog.Attr) context.Context {
	if existing, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		attrs = append(existing, attrs...)
	}
	return context.WithValue(ctx, slogFields, attrs)
}
