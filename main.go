package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/domust/fibonacci/api"
	"github.com/domust/fibonacci/internal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	telemetry, err := internal.NewTelemetry(ctx)
	if err != nil {
		log.Fatal(err)
	}

	slog.SetDefault(telemetry.Logger())
	metrics, err := internal.NewMetrics(telemetry.Meter())
	if err != nil {
		log.Fatal(err)
	}

	gs := grpc.NewServer(telemetry.ServerOption(), grpc.UnaryInterceptor(telemetry.UnaryInterceptor()))
	hs := health.NewServer()
	api.RegisterFibonacciServer(gs, internal.NewServer(metrics))
	grpc_health_v1.RegisterHealthServer(gs, hs)

	go func() {
		<-ctx.Done()
		cancel()
		hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		gs.GracefulStop()
	}()

	lis, err := net.Listen("tcp4", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting grpc server on %s\n", lis.Addr().String())
	if err := gs.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
