package internal

import (
	"context"

	"github.com/domust/fibonacci/api"
)

type Server struct {
	api.UnimplementedFibonacciServer
}

func (s *Server) GenerateSequence(_ context.Context, req *api.GenerateSequenceRequest) (*api.GenerateSequenceResponse, error) {
	return &api.GenerateSequenceResponse{}, nil
}
