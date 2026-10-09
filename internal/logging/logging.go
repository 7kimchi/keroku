// Package logging builds the JSON slog logger with secret redaction.
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

// New returns a JSON logger that redacts secrets from messages, keys and values.
func New(w io.Writer, level slog.Leveler, r *Redactor) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			return redactAttr(r, a)
		},
	}
	return slog.New(&handler{inner: slog.NewJSONHandler(w, opts), r: r})
}

func redactAttr(r *Redactor, a slog.Attr) slog.Attr {
	if sensitiveKey(a.Key) {
		return slog.String(a.Key, redacted)
	}
	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return slog.String(a.Key, r.String(v.String()))
	case slog.KindAny:
		switch x := v.Any().(type) {
		case error:
			return slog.String(a.Key, r.String(x.Error()))
		case fmt.Stringer:
			return slog.String(a.Key, r.String(x.String()))
		case []byte:
			return slog.String(a.Key, r.String(string(x)))
		default:
			return slog.String(a.Key, r.String(fmt.Sprintf("%+v", x)))
		}
	}
	return slog.Attr{Key: a.Key, Value: v}
}

// handler redacts the message, which ReplaceAttr never sees.
type handler struct {
	inner slog.Handler
	r     *Redactor
}

func (h *handler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h *handler) Handle(ctx context.Context, rec slog.Record) error {
	out := slog.NewRecord(rec.Time, rec.Level, h.r.String(rec.Message), rec.PC)
	rec.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(a)
		return true
	})
	return h.inner.Handle(ctx, out)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{inner: h.inner.WithAttrs(attrs), r: h.r}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{inner: h.inner.WithGroup(name), r: h.r}
}
