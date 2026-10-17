package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miekg/dns"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sneakybugs/corewarden/api/database"
	"github.com/sneakybugs/corewarden/api/services/health"
	"github.com/sneakybugs/corewarden/api/services/rest"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func toRRString(t *testing.T, rr string) string {
	parsedRR, err := dns.NewRR(rr)
	if err != nil {
		t.Fatalf("Expected error to be nil, got %v\n", err)
	}
	return parsedRR.String()
}

func TestCreateRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
}

func TestCreateCNAMEWhenCannot(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}

	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN CNAME bar.example.com."),
		Comment: "test",
	})
	if err != ErrCNAMEArgument {
		t.Fatalf("Expected an error '%v', got '%v'\n", ErrCNAMEArgument, err)
	}
}

func TestCreateRecordWhenCnameExists(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN CNAME bar.example.com."),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}

	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != ErrCNAMEArgument {
		t.Fatalf("Expected an error '%v', got '%v'\n", ErrCNAMEArgument, err)
	}
}

func TestReadRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	createResult, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
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
	expectedRR := toRRString(t, "foo 3600 IN A 127.0.0.1")
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
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected record not found error, got %v\n", err)
	}
}

func TestUpdateRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	createResult, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	updateResult, err := s.UpdateRecord(ctx, RecordUpdateParameters{
		ID:      createResult.ID,
		Zone:    createResult.Zone,
		RR:      createResult.RR,
		Comment: "changed",
	})
	if err != nil {
		t.Fatalf("failed to update record: %v\n", err)
	}
	if updateResult.Comment != "changed" {
		t.Fatalf("expected comment to be 'changed', got '%s'\n", updateResult.Comment)
	}
	readResult, err := s.ReadRecord(ctx, createResult.ID)
	if err != nil {
		t.Fatalf("failed to read record: %v\n", err)
	}
	if readResult.Comment != "changed" {
		t.Fatalf("expected comment to be 'test', got '%s'", readResult.Comment)
	}
}

func TestUpdateRecordNotFound(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.UpdateRecord(ctx, RecordUpdateParameters{
		ID:      1337,
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected record not found error, got %v\n", err)
	}
}

func TestDeleteRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	createResult, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	deleteResult, err := s.DeleteRecord(ctx, createResult.ID)
	if err != nil {
		t.Fatalf("failed to delete record: %v\n", err)
	}
	if deleteResult.Zone != "example.com." {
		t.Fatalf("expected zone to be 'example.com.', got '%s'", deleteResult.Zone)
	}
	expectedRR := toRRString(t, "foo 3600 IN A 127.0.0.1")
	if deleteResult.RR != expectedRR {
		t.Fatalf("expected RR to be '%s', got '%s'", expectedRR, deleteResult.RR)
	}
	if deleteResult.Comment != "test" {
		t.Fatalf("expected comment to be 'test', got '%s'", deleteResult.Comment)
	}
	_, err = s.ReadRecord(ctx, createResult.ID)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected error %v, got %v\n", ErrRecordNotFound, err)
	}
}

func TestDeleteRecordNotFound(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.DeleteRecord(ctx, 1337)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected record not found error, got %v\n", err)
	}
}

func TestListRecords(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "bar 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "foo.com.",
		RR:      toRRString(t, "@ 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	records, err := s.ListRecords(ctx, "example.com.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected records length to be 2, got %d", len(records))
	}
}

func TestListRecordsEmpty(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	records, err := s.ListRecords(ctx, "example.com.")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected records length to be 0, got %d", len(records))
	}

}

func TestResolveRecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
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

func TestResolveCNAMERecord(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN CNAME bar.example.com."),
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
	expectedAnswer := "foo.example.com.\t3600\tIN\tCNAME\tbar.example.com."
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
		RR:      toRRString(t, "foo 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo 3600 IN MX 10 mail.example.com."),
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
		RR:      toRRString(t, "@ 3600 IN A 127.0.0.1"),
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
		RR:      toRRString(t, "foo.bar.baz 3600 IN A 127.0.0.1"),
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

func TestResolveSubdomainWithDash(t *testing.T) {
	s, closer := createTestStorage()
	ctx := context.Background()
	defer closer(ctx)
	_, err := s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo-bar 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	res, err := s.Resolve(ctx, DNSQuestion{
		Name:  "foo-bar.example.com.",
		Qtype: dns.TypeA,
	})
	if err != nil {
		t.Fatalf("failed to resolve: %v\n", err)
	}
	if len(res.Answer) != 1 {
		t.Fatalf("expected answer length 1, got %d\n", len(res.Answer))
	}
	expectedAnswer := "foo-bar.example.com.\t3600\tIN\tA\t127.0.0.1"
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
		RR:      toRRString(t, "*.wildcard 3600 IN A 127.0.0.1"),
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
		RR:      toRRString(t, "*.wildcard 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "*.bar.wildcard 3600 IN A 0.0.0.0"),
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
		RR:      toRRString(t, "*.wildcard 3600 IN A 127.0.0.1"),
		Comment: "test",
	})
	if err != nil {
		t.Fatalf("failed to create record: %v\n", err)
	}
	_, err = s.CreateRecord(ctx, RecordCreateParameters{
		Zone:    "example.com.",
		RR:      toRRString(t, "foo.bar.wildcard 3600 IN A 0.0.0.0"),
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

	var handler *chi.Mux
	app := fx.New(
		fx.Supply(
			Options{
				ConnectionString: "postgres://development:development@localhost:5432/development?sslmode=disable",
			},
			zap.NewNop(),
		),
		fx.Provide(
			rest.NewMockService,
			health.NewReadinessChecks,
			NewService,
		),
		fx.Invoke(
			health.Register,
		),
		fx.Populate(
			&s,
			&handler,
		),
		fx.NopLogger,
	)
	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/-/readiness", nil))
		if w.Result().StatusCode == http.StatusOK {
			break
		}
		if time.Now().After(deadline) {
			panic("storage did not become ready")
		}
		time.Sleep(50 * time.Millisecond)
	}
	closer = func(ctx context.Context) {
		err := app.Stop(ctx)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}

	return
}

func TestStorageReadyAfterDatabaseBecomesAvailable(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve port: %v\n", err)
	}
	addr := l.Addr().String()
	_ = l.Close()

	var handler *chi.Mux
	app := fx.New(
		fx.Supply(
			Options{
				ConnectionString: fmt.Sprintf("postgres://development:development@%s/development?sslmode=disable", addr),
			},
			zap.NewNop(),
		),
		fx.Provide(
			rest.NewMockService,
			health.NewReadinessChecks,
			NewService,
		),
		fx.Invoke(
			health.Register,
			func(Storage) {},
		),
		fx.Populate(&handler),
		fx.NopLogger,
	)
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("expected start to succeed with database down, got %v\n", err)
	}
	defer func() { _ = app.Stop(ctx) }()

	readiness := func() int {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/-/readiness", nil))
		return w.Result().StatusCode
	}
	if code := readiness(); code != http.StatusInternalServerError {
		t.Fatalf("expected status code 500 while database is down, got %d\n", code)
	}

	time.Sleep(2 * time.Second)
	proxy, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to listen on %s: %v\n", addr, err)
	}
	defer func() { _ = proxy.Close() }()
	go forward(proxy, "localhost:5432")

	deadline := time.Now().Add(15 * time.Second)
	for readiness() != http.StatusOK {
		if time.Now().After(deadline) {
			t.Fatalf("expected readiness to become 200 after database came up\n")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func forward(l net.Listener, target string) {
	for {
		in, err := l.Accept()
		if err != nil {
			return
		}
		go func() {
			defer func() { _ = in.Close() }()
			out, err := net.Dial("tcp", target)
			if err != nil {
				return
			}
			defer func() { _ = out.Close() }()
			go func() { _, _ = io.Copy(out, in) }()
			_, _ = io.Copy(in, out)
		}()
	}
}
