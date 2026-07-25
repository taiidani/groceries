// Package telemetry configures the application's OpenTelemetry tracing
// pipeline.
//
// Everything is driven by the standard OTEL_* environment variables, so the
// application never hardcodes a destination:
//
//   - OTEL_TRACES_EXPORTER selects the exporter ("otlp" by default, "console"
//     to print spans locally, or "none" to disable exporting).
//   - OTEL_EXPORTER_OTLP_ENDPOINT / OTEL_EXPORTER_OTLP_PROTOCOL point the OTLP
//     exporter at a collector/agent (e.g. Grafana Alloy) that forwards traces
//     to a centralized backend such as Tempo.
//   - OTEL_SERVICE_NAME / OTEL_RESOURCE_ATTRIBUTES describe this service, e.g.
//     OTEL_RESOURCE_ATTRIBUTES=deployment.environment=prod,host.name=ntte-1.
//   - OTEL_TRACES_SAMPLER / OTEL_TRACES_SAMPLER_ARG control sampling; when unset
//     the SDK defaults to a parent-based always-on sampler.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init configures the global OpenTelemetry tracer provider and text-map
// propagator. It returns a shutdown function that flushes any buffered spans;
// callers should invoke it during graceful shutdown.
func Init(ctx context.Context) (func(context.Context) error, error) {
	// autoexport builds the span exporter from OTEL_TRACES_EXPORTER (otlp by
	// default; console and none also supported) and honours the standard
	// OTEL_EXPORTER_OTLP_* env vars for endpoint, protocol, and headers.
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("building otel span exporter: %w", err)
	}

	// The resource and sampler are left to the SDK, which builds them from
	// OTEL_SERVICE_NAME, OTEL_RESOURCE_ATTRIBUTES, and OTEL_TRACES_SAMPLER
	// (defaulting to ParentBased(AlwaysSample())).
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
