package internal

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/domust/fibonacci/api"
)

type Server struct {
	api.UnimplementedFibonacciServer
}

func (s *Server) GenerateSequence(ctx context.Context, req *api.GenerateSequenceRequest) (*api.GenerateSequenceResponse, error) {
	ctx, span := otel.GetTracerProvider().Tracer("github.com/domust/fibonacci/internal").Start(ctx, "GenerateSequence")
	defer span.End()

	return &api.GenerateSequenceResponse{}, nil
}
