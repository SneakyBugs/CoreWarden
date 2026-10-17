package resolver

import (
	"context"
	"fmt"
	"net"
	"slices"
	"testing"

	"github.com/miekg/dns"
	"github.com/sneakybugs/corewarden/api/resolver"
	grpcs "github.com/sneakybugs/corewarden/api/services/grpc"
	"github.com/sneakybugs/corewarden/api/services/logger"
	"github.com/sneakybugs/corewarden/api/services/storage"
	"github.com/sneakybugs/corewarden/api/services/zones"
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
	client, closer := createTestClient(t, storage.ResolveRecordNotFoundError)
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
	client, closer := createTestClient(t, storage.ResolveServerError)
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

func TestListZones(t *testing.T) {
	client, closer := createTestClient(t, nil, "example.com", "Example.NET.")
	defer closer(context.Background())

	resp, err := client.ListZones(context.Background(), &resolver.ListZonesRequest{})
	if err != nil {
		t.Fatalf("failed list zones request: %v", err)
	}
	if !slices.Equal(resp.Zones, []string{"example.com.", "example.net."}) {
		t.Errorf("unexpected zones: %v", resp.Zones)
	}
}

func TestListZonesEmpty(t *testing.T) {
	client, closer := createTestClient(t, nil)
	defer closer(context.Background())

	resp, err := client.ListZones(context.Background(), &resolver.ListZonesRequest{})
	if err != nil {
		t.Fatalf("failed list zones request: %v", err)
	}
	if len(resp.Zones) != 0 {
		t.Errorf("expected no zones, got %v", resp.Zones)
	}
}

func TestResolveIgnoresConfiguredZones(t *testing.T) {
	client, closer := createTestClient(t, nil, "example.net.")
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
	if len(resp.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(resp.Answer))
	}
}

func createTestClient(t *testing.T, returnError error, configuredZones ...string) (resolver.ResolverClient, func(context.Context)) {
	lis := bufconn.Listen(10 * 1024 * 1024)
	var store storage.Storage
	app := fx.New(
		fx.Supply(
			storage.MockStorageOptions{
				ReturnError: returnError,
			},
			logger.Options{
				DevelopmentMode: true,
			},
			zones.Options{
				Zones: configuredZones,
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

	conn, err := grpc.NewClient(
		"127.0.0.1",
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
		_ = lis.Close()
	}
}
