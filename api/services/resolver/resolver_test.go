package resolver

import (
	"context"
	"database/sql"
	"net"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/database"
	"git.houseofkummer.com/lior/home-dns/api/resolver"
	grpcs "git.houseofkummer.com/lior/home-dns/api/services/grpc"
	"git.houseofkummer.com/lior/home-dns/api/services/logger"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"github.com/miekg/dns"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestResolve(t *testing.T) {
	client, closer, store := createTestClient(t)
	// Why does it not function?
	defer closer(context.Background())

	_, err := store.CreateRecord(context.Background(), storage.RecordCreateParameters{
		Zone:    "example.com",
		RR:      "foo IN A 192.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("expected to create record with no error, got %v", err)
	}

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
	closer(context.Background())
}

func createTestClient(t *testing.T) (resolver.ResolverClient, func(context.Context), *storage.Storage) {
	migrations := database.GetMigrations()
	db, err := sql.Open("pgx", "postgres://development:development@localhost:5432/development?sslmode=disable")
	if err != nil {
		panic(err)
	}
	_, err = migrate.Exec(db, "postgres", migrations, migrate.Down)
	if err != nil {
		panic(err)
	}

	lis := bufconn.Listen(10 * 1024 * 1024)
	var store *storage.Storage
	app := fx.New(
		fx.Supply(
			storage.Options{
				ConnectionString: "postgres://development:development@localhost:5432/development?sslmode=disable",
			},
		),
		fx.Provide(
			grpcs.NewService,
			logger.NewService,
			storage.NewService,
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
		app.Stop(ctx)
		lis.Close()
	}, store
}
