package main

import (
	"context"
	"log"
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
	lis, err := net.Listen("tcp4", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s\n", lis.Addr().String())

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	telemetry, err := internal.NewTelemetry(ctx)
	if err != nil {
		log.Fatal(err)
	}

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

	log.Println("starting grpc server")
	if err := gs.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
