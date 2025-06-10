package internal

import (
	"context"

	"github.com/domust/fibonacci/api"
)

// Server implements the [api.FibonacciServer] interface.
type Server struct {
	api.UnimplementedFibonacciServer

	metrics *Metrics
}

// NewServer returns server configured with instrumentation.
func NewServer(metrics *Metrics) *Server {
	return &Server{
		metrics: metrics,
	}
}

// GenerateSequence is part of the [api.FibonacciServer] interface.
func (s *Server) GenerateSequence(ctx context.Context, req *api.GenerateSequenceRequest) (*api.GenerateSequenceResponse, error) {
	s.metrics.counter.Add(ctx, 1)

	return &api.GenerateSequenceResponse{}, nil
}
