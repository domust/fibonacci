package internal

import (
	"context"
	"iter"

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

	seq := make([]uint64, 0, req.GetLength())
	for num := range fibonacci(req.GetLength()) {
		seq = append(seq, num)
	}

	return &api.GenerateSequenceResponse{Sequence: seq}, nil
}

func fibonacci(length uint32) iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		var a, b uint64

		for i := range uint64(length) {
			if i <= 1 {
				a += i
			}
			a, b = b, a+b

			if !yield(b) {
				return
			}
		}
	}
}
