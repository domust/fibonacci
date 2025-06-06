package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/domust/fibonacci/api"
	"github.com/domust/fibonacci/internal"
)

func main() {
	lis, err := net.Listen("tcp4", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s\n", lis.Addr().String())

	srv := grpc.NewServer()
	api.RegisterFibonacciServer(srv, &internal.Server{})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancel()
		srv.GracefulStop()
	}()

	log.Println("starting grpc server")
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
