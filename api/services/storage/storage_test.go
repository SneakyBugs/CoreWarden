package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"git.houseofkummer.com/lior/home-dns/api/database"
	"github.com/miekg/dns"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/fx"
)

func TestCreateRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
}

func TestReadRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	createResult, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	readResult, err := s.ReadRecord(ctx, createResult.ID)
	if err != nil {
		t.Fatalf("failed to read record: %v\n", err)
	}
	if readResult.Zone != "example.com." {
		t.Fatalf("expected zone to be 'example.com.', got '%s'", readResult.Zone)
	}
	expectedRR := "foo 3600 IN A 127.0.0.1"
	if readResult.RR != expectedRR {
		t.Fatalf("expected RR to be '%s', got '%s'", expectedRR, readResult.RR)
	}
	if readResult.Comment != "test" {
		t.Fatalf("expected comment to be 'test', got '%s'", readResult.Comment)
	}
}

func TestReadRecordNotFound(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.ReadRecord(ctx, 1337)
	if !errors.Is(err, RecordNotFoundError) {
		t.Fatalf("expected record not found error, got %v\n", err)
	}
}

func TestResolveRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "foo.example.com.\t3600\tIN\tA\t127.0.0.1"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func TestResolveRecordQtype(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo 3600 IN CNAME foo.example.com",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "foo.example.com.\t3600\tIN\tA\t127.0.0.1"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func TestResolveRecordAt(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "@ 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "example.com.\t3600\tIN\tA\t127.0.0.1"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func TestResolveSubdomain(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo.bar.baz 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo.bar.baz.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "foo.bar.baz.example.com.\t3600\tIN\tA\t127.0.0.1"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func TestResolveWildcard(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "*.wildcard 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo.bar.wildcard.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "foo.bar.wildcard.example.com.\t3600\tIN\tA\t127.0.0.1"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func TestResolveWildcardPrecedence(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "*.wildcard 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "*.bar.wildcard 3600 IN A 0.0.0.0",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo.bar.wildcard.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "foo.bar.wildcard.example.com.\t3600\tIN\tA\t0.0.0.0"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func TestResolvePrecedenceOverWildcard(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "*.wildcard 3600 IN A 127.0.0.1",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      "foo.bar.wildcard 3600 IN A 0.0.0.0",
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo.bar.wildcard.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	if !strings.HasSuffix(res.Answer[0], "0.0.0.0") {
		t.Fatalf("expected answer to be 0.0.0.0, got %s\n", res.Answer[0])
	}
	expectedAnswer := "foo.bar.wildcard.example.com.\t3600\tIN\tA\t0.0.0.0"
	if res.Answer[0] != expectedAnswer {
		t.Fatalf("expected answer to be '%s', got '%s'", expectedAnswer, res.Answer[0])
	}
}

func createTestStorage() (s Storage, closer func(context.Context)) {
	migrations := database.GetMigrations()
	db, err := sql.Open("pgx", "postgres://development:development@localhost:5432/development?sslmode=disable")
	if err != nil {
		panic(err)
	}
	_, err = migrate.Exec(db, "postgres", migrations, migrate.Down)
	if err != nil {
		panic(err)
	}

	app := fx.New(
		fx.Supply(
			Options{
				ConnectionString: "postgres://development:development@localhost:5432/development?sslmode=disable",
			},
		),
		fx.Provide(
			NewService,
		),
		fx.Populate(
			&s,
		),
	)
	go func() {
		app.Run()
	}()
	closer = func(ctx context.Context) {
		err := app.Stop(ctx)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}

	return
}
