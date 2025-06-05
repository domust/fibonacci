package main

import (
	"log"
	"net"

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

	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
