package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/domust/fibonacci/api"
	"github.com/domust/fibonacci/internal"
	rpc "github.com/domust/fibonacci/internal/grpc"
	"github.com/domust/fibonacci/internal/telemetry"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	tel, err := telemetry.NewTelemetry(ctx)
	if err != nil {
		log.Fatal(err)
	}

	slog.SetDefault(tel.Logger()) // comment out in order to debug startup failures locally
	metrics, err := telemetry.NewMetrics(tel.Meter())
	if err != nil {
		log.Fatal(err)
	}

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal(err)
	}

	gs := rpc.NewServer(tel, validator)
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
	go func() {
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	proxy := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = api.RegisterFibonacciHandlerFromEndpoint(ctx, proxy, "0.0.0.0:8080", opts)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("starting grpc proxy on 0.0.0.0:8081")
	if err := http.ListenAndServe(":8081", proxy); err != nil {
		log.Fatal(err)
	}
}
