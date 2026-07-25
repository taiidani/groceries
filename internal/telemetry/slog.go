package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler wraps a slog.Handler and enriches each record with the trace_id
// and span_id of the span found in the record's context. Combined with a
// Grafana Loki-to-Tempo derived field on trace_id, this lets you jump from a
// log line straight to the trace that emitted it.
//
// Correlation only occurs when a log is emitted with a context carrying an
// active span (e.g. slog.InfoContext(ctx, ...)); the context-free helpers such
// as slog.Info pass a background context and therefore carry no span.
type TraceHandler struct {
	slog.Handler
}

// NewTraceHandler wraps the given handler so emitted records are annotated with
// trace correlation attributes.
func NewTraceHandler(h slog.Handler) *TraceHandler {
	return &TraceHandler{Handler: h}
}

func (h *TraceHandler) Handle(ctx context.Context, record slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, record)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}
