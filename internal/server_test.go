package internal

import (
	"context"
	"log"
	"net"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/domust/fibonacci/api"
	rpc "github.com/domust/fibonacci/internal/grpc"
)

func TestServer(t *testing.T) {
	server := func() (api.FibonacciClient, func()) {
		validator, err := protovalidate.New()
		if err != nil {
			log.Fatal(err)
		}

		s := rpc.NewServer(nil, validator)
		api.RegisterFibonacciServer(s, &Server{})

		lis := bufconn.Listen(1024 * 1024)
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatal(err)
			}
		}()

		conn, err := grpc.NewClient("passthrough://", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatal(err)
		}

		return api.NewFibonacciClient(conn), s.Stop
	}

	client, stop := server()
	t.Cleanup(stop)

	t.Run("input validation", func(t *testing.T) {
		resp, err := client.GenerateSequence(context.Background(), &api.GenerateSequenceRequest{})
		require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
		require.Nil(t, resp)
	})

	t.Run("overflow prevention", func(t *testing.T) {
		resp, err := client.GenerateSequence(context.Background(), &api.GenerateSequenceRequest{Length: 95})
		require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
		require.Nil(t, resp)
	})

	tests := map[string]struct {
		input  uint32
		output []uint64
	}{
		"single digit": {
			input:  1,
			output: []uint64{0},
		},
		"two digits": {
			input:  2,
			output: []uint64{0, 1},
		},
		"three digits": {
			input:  3,
			output: []uint64{0, 1, 1},
		},
		"four digits": {
			input:  4,
			output: []uint64{0, 1, 1, 2},
		},
		"five digits": {
			input:  5,
			output: []uint64{0, 1, 1, 2, 3},
		},
		"ten digits": {
			input:  10,
			output: []uint64{0, 1, 1, 2, 3, 5, 8, 13, 21, 34},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := client.GenerateSequence(context.Background(), &api.GenerateSequenceRequest{Length: data.input})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, data.output, resp.Sequence)
		})
	}
}
