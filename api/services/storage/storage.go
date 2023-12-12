package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.houseofkummer.com/lior/home-dns/api/database/queries"
	"github.com/miekg/dns"
)

type Storage interface {
	Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error)
	CreateRecord(ctx context.Context, p RecordCreateParameters) (Record, error)
	// ReadRecord(ctx context.Context, id int) (Record, error)
}

type PostgresStorage struct {
	queries *queries.Queries
}

func (s *PostgresStorage) Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error) {
	r, err := s.queries.ResolveRecord(ctx, queries.ResolveRecordParams{
		Name: q.Name,
		Type: int32(q.Qtype),
	})
	if err != nil {
		return DNSResponse{}, err
	}
	if len(r) != 0 {
		answer := make([]string, len(r))
		for i, record := range r {
			rr, err := replaceName(record.Content, q.Name)
			if err != nil {
				return DNSResponse{}, fmt.Errorf("failed parsing record")
			}
			answer[i] = rr
		}
		return DNSResponse{Answer: answer}, nil
	}
	labels := dns.SplitDomainName(q.Name)
	subdomains := make([]string, len(labels))
	currentSubdomain := ""
	for i := len(labels) - 1; 0 < i; i-- {
		currentSubdomain = labels[i] + "." + currentSubdomain
		subdomains[i] = currentSubdomain
	}
	r, err = s.queries.ResolveWildcardRecord(ctx, queries.ResolveWildcardRecordParams{
		Names: subdomains,
		Type:  int32(q.Qtype),
	})
	if err != nil {
		return DNSResponse{}, err
	}
	if len(r) == 0 {
		return DNSResponse{}, fmt.Errorf("no records found")
	}
	maxLength := 0
	answer := []string{}
	for _, record := range r {
		if maxLength < len(record.Name) {
			maxLength = len(record.Name)
			rr, err := replaceName(record.Content, q.Name)
			if err != nil {
				return DNSResponse{}, fmt.Errorf("failed parsing record")
			}
			answer = []string{rr}
		} else if maxLength == len(record.Name) {
			rr, err := replaceName(record.Content, q.Name)
			if err != nil {
				return DNSResponse{}, fmt.Errorf("failed parsing record")
			}
			answer = append(answer, rr)
		}
	}
	return DNSResponse{Answer: answer}, nil
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
		ID:      int(r.ID),
		Zone:    r.Zone,
		RR:      r.Content,
		Comment: r.Comment,
	}, nil
}

type RecordCreateParameters struct {
	Zone    string
	RR      string
	Comment string
}

type Record struct {
	ID        int
	Zone      string
	RR        string
	Comment   string
	CreatedAt time.Time
	UpdatedOn time.Time
}
