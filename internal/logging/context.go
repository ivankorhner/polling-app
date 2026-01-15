package logging

import (
	"context"
	"log/slog"
)

type ctxKey string

const slogFields ctxKey = "slogFields"

// ContextHandler is a slog.Handler that extracts attributes from context
type ContextHandler struct {
	slog.Handler
}

// Handle implements slog.Handler and adds context attributes to the record
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, attr := range attrs {
			r.AddAttrs(attr)
		}
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new ContextHandler with the given attributes
func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new ContextHandler with the given group name
func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{Handler: h.Handler.WithGroup(name)}
}

// AppendCtx appends slog attributes to the context
func AppendCtx(ctx context.Context, attrs ...slog.Attr) context.Context {
	if existing, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		attrs = append(existing, attrs...)
	}
	return context.WithValue(ctx, slogFields, attrs)
}
