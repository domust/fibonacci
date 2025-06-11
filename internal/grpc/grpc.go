// Package grpc ensures consistency between test and non-test grpc servers.
package grpc

import (
	"buf.build/go/protovalidate"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"

	"github.com/domust/fibonacci/internal/telemetry"
)

// NewServer is a wrapper around [google.golang.org/grpc.NewServer] to ensure that
// server configuration is identical between production and test servers.
func NewServer(
	telemetry *telemetry.Telemetry,
	validator protovalidate.Validator,
) *grpc.Server {
	var opts []grpc.ServerOption
	if telemetry != nil {
		opts = append(opts, telemetry.ServerOption())
	}

	var interceptors []grpc.UnaryServerInterceptor
	if telemetry != nil {
		interceptors = append(interceptors, telemetry.UnaryInterceptor())
	}
	interceptors = append(interceptors, middleware.UnaryServerInterceptor(validator))
	opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))

	return grpc.NewServer(opts...)
}
