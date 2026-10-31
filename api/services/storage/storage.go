package storage

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/miekg/dns"
	"github.com/sneakybugs/corewarden/api/database/queries"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Storage interface {
	Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error)
	CreateRecord(ctx context.Context, p RecordCreateParameters) (Record, error)
	ReadRecord(ctx context.Context, id int) (Record, error)
	UpdateRecord(ctx context.Context, p RecordUpdateParameters) (Record, error)
	DeleteRecord(ctx context.Context, id int) (Record, error)
	ListRecords(ctx context.Context, zone string) ([]Record, error)
}

var ErrRecordNotFound = errors.New("record not found")
var ErrServer = errors.New("server error")
var ErrCNAMEArgument = errors.New("CNAME must be the only record at a node")

var ResolveServerError = status.Error(
	codes.Internal,
	"internal server error",
)
var ResolveRecordNotFoundError = status.Error(
	codes.NotFound,
	"record not found",
)

// noDataSOA is the SOA without owner for NODATA responses. TTL and MINIMUM
// are 0 so negative answers are not cached.
const noDataSOA = "0\tIN\tSOA\tns.icann.org. noc.dns.icann.org. 1 7200 3600 1209600 0"

type PostgresStorage struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func (s *PostgresStorage) Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error) {
	records, err := s.queries.ResolveRecord(ctx, q.Name)
	if err != nil {
		return DNSResponse{}, ResolveServerError
	}
	if len(records) == 0 {
		records, err = s.queries.ResolveWildcardRecord(ctx, enclosingNames(q.Name))
		if err != nil {
			return DNSResponse{}, ResolveServerError
		}
		records = closest(records)
	}
	if len(records) == 0 {
		return DNSResponse{}, ResolveRecordNotFoundError
	}
	answer := []string{}
	for _, record := range records {
		if record.Type != int32(q.Qtype) && record.Type != int32(dns.TypeCNAME) {
			continue
		}
		rr, err := replaceName(record.Content, q.Name)
		if err != nil {
			return DNSResponse{}, ResolveServerError
		}
		answer = append(answer, rr)
	}
	if len(answer) == 0 {
		return DNSResponse{NS: []string{records[0].Zone + "\t" + noDataSOA}}, nil
	}
	return DNSResponse{Answer: answer}, nil
}

func enclosingNames(name string) []string {
	labels := dns.SplitDomainName(name)
	names := make([]string, 0, len(labels))
	for i := 1; i < len(labels); i++ {
		names = append(names, strings.Join(labels[i:], ".")+".")
	}
	return names
}

func closest(records []queries.Record) []queries.Record {
	longest := 0
	for _, r := range records {
		longest = max(longest, len(r.Name))
	}
	return slices.DeleteFunc(records, func(r queries.Record) bool {
		return len(r.Name) != longest
	})
}

func replaceName(rr string, newName string) (string, error) {
	parsed, err := dns.NewRR(rr)
	if err != nil {
		return "", err
	}
	parsed.Header().Name = newName
	return parsed.String(), nil
}

type DNSQuestion struct {
	Name  string
	Qtype uint16
}

type DNSResponse struct {
	Answer []string
	NS     []string
	Extra  []string
}

func (s *PostgresStorage) CreateRecord(ctx context.Context, p RecordCreateParameters) (Record, error) {
	rr, err := dns.NewRR(p.RR)
	if err != nil {
		return Record{}, fmt.Errorf("failed to parse RR: %v", err)
	}

	zoneFqdn := dns.Fqdn(p.Zone)
	rrName := dns.Fqdn(rr.Header().Name)
	fullName := rrName + zoneFqdn
	isWildcard := false
	if rrName == "." {
		fullName = zoneFqdn
	} else if strings.HasPrefix(fullName, "*.") {
		fullName = fullName[2:]
		isWildcard = true
	}

	// RFC 1034 section 3.6.2: "If a CNAME RR is present at a node, no other data should be present"
	tx, err := s.pool.Begin(ctx)
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err != nil {
		return Record{}, err
	}
	if _, ok := rr.(*dns.CNAME); ok {
		anyExist, err := s.queries.AnyRecordsExistAtNode(ctx, queries.AnyRecordsExistAtNodeParams{
			Zone: zoneFqdn,
			Name: fullName,
		})
		if err != nil {
			return Record{}, err
		}
		if anyExist {
			return Record{}, ErrCNAMEArgument
		}
	} else {
		cnameExists, err := s.queries.CNAMERecordExistsAtNode(ctx, queries.CNAMERecordExistsAtNodeParams{
			Zone: zoneFqdn,
			Name: fullName,
		})
		if err != nil {
			return Record{}, err
		}
		if cnameExists {
			return Record{}, ErrCNAMEArgument
		}
	}

	r, err := s.queries.CreateRecord(ctx, queries.CreateRecordParams{
		Zone:       zoneFqdn,
		Content:    p.RR,
		Name:       fullName,
		IsWildcard: isWildcard,
		Type:       int32(rr.Header().Rrtype),
		Comment:    p.Comment,
	})
	if err != nil {
		return Record{}, fmt.Errorf("faild to create record: %v", err)
	}
	return Record{
		ID:         int(r.ID),
		Zone:       r.Zone,
		RR:         r.Content,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt.Time,
		ModifiedOn: r.ModifiedOn.Time,
	}, tx.Commit(ctx)
}
func (s *PostgresStorage) ReadRecord(ctx context.Context, id int) (Record, error) {
	r, err := s.queries.ReadRecord(ctx, int32(id))
	if err != nil {
		return Record{}, ErrRecordNotFound
	}
	return Record{
		ID:         int(r.ID),
		Zone:       r.Zone,
		RR:         r.Content,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt.Time,
		ModifiedOn: r.ModifiedOn.Time,
	}, nil
}

func (s *PostgresStorage) UpdateRecord(ctx context.Context, p RecordUpdateParameters) (Record, error) {
	rr, err := dns.NewRR(p.RR)
	if err != nil {
		return Record{}, fmt.Errorf("failed to parse RR: %v", err)
	}

	zoneFqdn := dns.Fqdn(p.Zone)
	rrName := dns.Fqdn(rr.Header().Name)
	fullName := rrName + zoneFqdn
	isWildcard := false
	if rrName == "." {
		fullName = zoneFqdn
	} else if strings.HasPrefix(fullName, "*.") {
		fullName = fullName[2:]
		isWildcard = true
	}

	r, err := s.queries.UpdateRecord(ctx, queries.UpdateRecordParams{
		ID:         int32(p.ID),
		Zone:       zoneFqdn,
		Content:    p.RR,
		Name:       fullName,
		IsWildcard: isWildcard,
		Type:       int32(rr.Header().Rrtype),
		Comment:    p.Comment,
	})
	if err != nil {
		return Record{}, ErrRecordNotFound
	}
	return Record{
		ID:         int(r.ID),
		Zone:       r.Zone,
		RR:         r.Content,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt.Time,
		ModifiedOn: r.ModifiedOn.Time,
	}, nil
}

func (s *PostgresStorage) DeleteRecord(ctx context.Context, id int) (Record, error) {
	r, err := s.queries.DeleteRecord(ctx, int32(id))
	if err != nil {
		return Record{}, ErrRecordNotFound
	}
	return Record{
		ID:         int(r.ID),
		Zone:       r.Zone,
		RR:         r.Content,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt.Time,
		ModifiedOn: r.ModifiedOn.Time,
	}, nil
}

func (s *PostgresStorage) ListRecords(ctx context.Context, zone string) ([]Record, error) {
	r, err := s.queries.ListRecords(ctx, zone)
	if err != nil {
		return []Record{}, ErrServer
	}
	records := make([]Record, len(r))
	for i, record := range r {
		records[i] = Record{
			ID:         int(record.ID),
			Zone:       record.Zone,
			RR:         record.Content,
			Comment:    record.Comment,
			CreatedAt:  record.CreatedAt.Time,
			ModifiedOn: record.ModifiedOn.Time,
		}
	}
	return records, nil
}

type RecordCreateParameters struct {
	Zone    string
	RR      string
	Comment string
}

type RecordUpdateParameters struct {
	ID      int
	Zone    string
	RR      string
	Comment string
}

type Record struct {
	ID         int
	Zone       string
	RR         string
	Comment    string
	CreatedAt  time.Time
	ModifiedOn time.Time
}
