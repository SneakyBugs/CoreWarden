package resolver

import (
	"context"
	"fmt"
	"net"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/resolver"
	grpcs "git.houseofkummer.com/lior/home-dns/api/services/grpc"
	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestResolve(t *testing.T) {
	client, closer := createTestClient(t, nil)
	defer closer(context.Background())

	resp, err := client.Resolve(
		context.Background(),
		&resolver.Question{
			Name:  "foo.example.com.",
			Qtype: uint32(dns.TypeA),
		},
	)
	if err != nil {
		t.Fatalf("failed resolve request: %v", err)
	}
	if resp == nil {
		t.Fatalf("response should not be nil")
	}
	if len(resp.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(resp.Answer))
	}
}

func TestResolveNotFound(t *testing.T) {
	client, closer := createTestClient(t, RecordNotFoundError)
	defer closer(context.Background())

	_, err := client.Resolve(
		context.Background(),
		&resolver.Question{
			Name:  "foo.example.com.",
			Qtype: uint32(dns.TypeA),
		},
	)
	if c := status.Convert(err).Code(); c != codes.NotFound {
		t.Fatalf("expected code %d, got %d", codes.NotFound, c)
	}
}

func TestResolveServerError(t *testing.T) {
	client, closer := createTestClient(t, ServerError)
	defer closer(context.Background())

	_, err := client.Resolve(
		context.Background(),
		&resolver.Question{
			Name:  "foo.example.com.",
			Qtype: uint32(dns.TypeA),
		},
	)
	if c := status.Convert(err).Code(); c != codes.Internal {
		t.Fatalf("expected code %d, got %d", codes.Internal, c)
	}
}

func createTestClient(t *testing.T, returnError error) (resolver.ResolverClient, func(context.Context)) {
	lis := bufconn.Listen(10 * 1024 * 1024)
	var store storage.Storage
	app := fx.New(
		fx.Supply(
			storage.MockStorageOptions{
				ReturnError: returnError,
			},
		),
		fx.Provide(
			grpcs.NewService,
			logger.NewService,
			storage.NewMockService,
			func() net.Listener {
				return lis
			},
		),
		fx.Invoke(
			Register,
		),
		fx.Populate(
			&store,
		),
	)
	go func() {
		app.Run()
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"",
		grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				return lis.Dial()
			},
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed creating gRPC client: %v", err)
	}

	return resolver.NewResolverClient(conn), func(ctx context.Context) {
		err := app.Stop(ctx)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		lis.Close()
	}
}
