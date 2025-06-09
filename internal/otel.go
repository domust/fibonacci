package internal

import (
	"context"
	"net"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	tracing "google.golang.org/grpc/experimental/opentelemetry"
	"google.golang.org/grpc/stats/opentelemetry"
)

func WithTracing(ctx context.Context) grpc.ServerOption {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithDialOption(grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	})))
	if err != nil {
		panic(err)
	}

	return opentelemetry.ServerOption(opentelemetry.Options{
		TraceOptions: tracing.TraceOptions{
			TracerProvider:    newTracerProvider(exporter),
			TextMapPropagator: propagation.TraceContext{},
		},
	})
}

func newTracerProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	rsc, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(semconv.ServiceName("Fibonacci")),
	)
	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(rsc),
	)
	otel.SetTracerProvider(tp)

	return tp
}
