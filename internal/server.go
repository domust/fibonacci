package internal

import (
	"context"

	"github.com/domust/fibonacci/api"
	"github.com/domust/fibonacci/internal/telemetry"
)

// Server implements the [api.FibonacciServer] interface.
type Server struct {
	api.UnimplementedFibonacciServer

	metrics *telemetry.Metrics
}

// NewServer returns server configured with instrumentation.
func NewServer(metrics *telemetry.Metrics) *Server {
	return &Server{
		metrics: metrics,
	}
}

// GenerateSequence is part of the [api.FibonacciServer] interface.
func (s *Server) GenerateSequence(ctx context.Context, req *api.GenerateSequenceRequest) (*api.GenerateSequenceResponse, error) {
	// TODO: reenable later
	//s.metrics.counter.Add(ctx, 1)

	return &api.GenerateSequenceResponse{}, nil
}
