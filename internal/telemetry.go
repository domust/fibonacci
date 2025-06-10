package internal

import (
	"context"
	"net"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	tracing "google.golang.org/grpc/experimental/opentelemetry"
	"google.golang.org/grpc/stats/opentelemetry"
)

// Telemetry encapsulates dependencies required to produce telemetry signals.
type Telemetry struct {
	provider   trace.TracerProvider
	propagator propagation.TextMapPropagator
}

// ServerOption is required to start a span when the server's Recv method is called.
func (t *Telemetry) ServerOption() grpc.ServerOption {
	return opentelemetry.ServerOption(opentelemetry.Options{
		TraceOptions: tracing.TraceOptions{
			TracerProvider:    t.provider,
			TextMapPropagator: t.propagator,
		},
	})
}

// UnaryInterceptor is required to create a method specific span from the parent (Recv) span.
func (t *Telemetry) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, span := t.provider.Tracer("github.com/domust/fibonacci").Start(ctx, info.FullMethod)
		defer span.End()

		return handler(ctx, req)
	}
}

// NewTelemetry is used to provision dependencies required for exporting telemetry signals.
func NewTelemetry(ctx context.Context) (*Telemetry, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithDialOption(grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	})))
	if err != nil {
		return nil, err
	}

	rsc, err := newResource(ctx)
	if err != nil {
		return nil, err
	}

	provider, err := newTracerProvider(exporter, rsc)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		provider:   provider,
		propagator: propagation.TraceContext{},
	}, nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	rsc, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithProcess(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceName("Fibonacci"),
			semconv.ServiceNamespace("fibonacci"),
		),
	)
	if err != nil {
		return nil, err
	}

	return resource.Merge(resource.Default(), rsc)
}

func newTracerProvider(exp sdktrace.SpanExporter, rsc *resource.Resource) (*sdktrace.TracerProvider, error) {
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(rsc),
	), nil
}
