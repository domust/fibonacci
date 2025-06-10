package internal

import (
	"context"
	"net"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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
	traces     trace.TracerProvider
	propagator propagation.TextMapPropagator
	metrics    metric.MeterProvider
}

// ServerOption is required to start a span when the server's Recv method is called.
func (t *Telemetry) ServerOption() grpc.ServerOption {
	return opentelemetry.ServerOption(opentelemetry.Options{
		TraceOptions: tracing.TraceOptions{
			TracerProvider:    t.traces,
			TextMapPropagator: t.propagator,
		},
	})
}

// UnaryInterceptor is required to create a method specific span from the parent (Recv) span.
func (t *Telemetry) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, span := t.traces.Tracer("github.com/domust/fibonacci").Start(ctx, info.FullMethod)
		defer span.End()

		return handler(ctx, req)
	}
}

func (t *Telemetry) Meter() metric.Meter {
	return t.metrics.Meter("github.com/domust/fibonacci")
}

// NewTelemetry is used to provision dependencies required for exporting telemetry signals.
func NewTelemetry(ctx context.Context) (*Telemetry, error) {
	traces, err := otlptracegrpc.New(ctx, otlptracegrpc.WithDialOption(grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	})))
	if err != nil {
		return nil, err
	}

	metrics, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	rsc, err := newResource(ctx)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		traces:     newTracerProvider(traces, rsc),
		propagator: propagation.TraceContext{},
		metrics:    newMeterProvider(metrics, rsc),
	}, nil
}

// Metrics encapsulates all metrics fox export.
type Metrics struct {
	counter metric.Int64Counter
}

// NewMetrics creates metrics from a given meter.
func NewMetrics(meter metric.Meter) (*Metrics, error) {
	counter, err := meter.Int64Counter("fibonacci.requests.count")
	if err != nil {
		return nil, err
	}

	return &Metrics{
		counter: counter,
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

func newTracerProvider(exp sdktrace.SpanExporter, rsc *resource.Resource) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(rsc),
		sdktrace.WithBatcher(exp),
	)
}

func newMeterProvider(exp sdkmetric.Exporter, rsc *resource.Resource) *sdkmetric.MeterProvider {
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(rsc),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exp),
		),
	)
}
