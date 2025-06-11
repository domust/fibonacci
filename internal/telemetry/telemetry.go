// Package telemetry implements sending of telemetry signals, such as traces, metrics, and logs.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	tracing "google.golang.org/grpc/experimental/opentelemetry"
	"google.golang.org/grpc/stats/opentelemetry"
)

const scope = "github.com/domust/fibonacci"

// Telemetry encapsulates dependencies required to produce telemetry signals.
type Telemetry struct {
	traces     trace.TracerProvider
	propagator propagation.TextMapPropagator
	metrics    metric.MeterProvider
	logs       log.LoggerProvider
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
		ctx, span := t.traces.Tracer(scope).Start(ctx, info.FullMethod)
		defer span.End()

		return handler(ctx, req)
	}
}

// Meter returns a new meter for dependency injection.
func (t *Telemetry) Meter() metric.Meter {
	return t.metrics.Meter(scope)
}

// Logger returns standard library's structured logger configured with telemetry.
func (t *Telemetry) Logger() *slog.Logger {
	return otelslog.NewLogger(scope, otelslog.WithLoggerProvider(t.logs))
}

// NewTelemetry is used to provision dependencies required for exporting telemetry signals.
func NewTelemetry(ctx context.Context) (*Telemetry, error) {
	traces, err := otlptracegrpc.New(ctx, otlptracegrpc.WithDialOption(grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	})))
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}

	metrics, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("metric exporter: %w", err)
	}

	logs, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("log exporter: %w", err)
	}

	rsc, err := newResource(ctx)
	if err != nil {
		return nil, err
	}

	return &Telemetry{
		traces:     newTracerProvider(traces, rsc),
		propagator: propagation.TraceContext{},
		metrics:    newMeterProvider(metrics, rsc),
		logs:       newLoggerProvider(logs, rsc),
	}, nil
}

// Metrics encapsulates all metrics fox export.
type Metrics struct {
	counter metric.Int64Counter
}

// Inc adds to the API request counter.
func (m *Metrics) Inc(ctx context.Context) {
	if m == nil {
		return
	}
	m.counter.Add(ctx, 1)
}

// NewMetrics creates metrics from a given meter.
func NewMetrics(meter metric.Meter) (*Metrics, error) {
	counter, err := meter.Int64Counter("fibonacci.requests.count")
	if err != nil {
		return nil, fmt.Errorf("counter: %w", err)
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
		return nil, fmt.Errorf("resource: %w", err)
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

func newLoggerProvider(exp sdklog.Exporter, rsc *resource.Resource) *sdklog.LoggerProvider {
	return sdklog.NewLoggerProvider(
		sdklog.WithResource(rsc),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exp),
		),
	)
}
