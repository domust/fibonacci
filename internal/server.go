package internal

import (
	"context"

	"github.com/domust/fibonacci/api"
)

// Server implements the [api.FibonacciServer] interface.
type Server struct {
	api.UnimplementedFibonacciServer
}

// GenerateSequence is part of the [api.FibonacciServer] interface.
func (s *Server) GenerateSequence(ctx context.Context, req *api.GenerateSequenceRequest) (*api.GenerateSequenceResponse, error) {
	return &api.GenerateSequenceResponse{}, nil
}
