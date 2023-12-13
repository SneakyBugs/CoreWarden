package storage

import (
	"context"
	"errors"
	"time"
)

type MockStorage struct {
	nextID  int
	records []Record
}

func (s *MockStorage) Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error) {
	return DNSResponse{
		Answer: []string{q.Name + "\tIN\tA\t127.0.0.1"},
	}, nil
}

func (s *MockStorage) CreateRecord(ctx context.Context, p RecordCreateParameters) (Record, error) {
	record := Record{
		ID:        s.nextID,
		Zone:      p.Zone,
		RR:        p.RR,
		Comment:   p.Comment,
		CreatedAt: time.Now(),
		UpdatedOn: time.Now(),
	}
	s.records = append(s.records, record)
	s.nextID++
	return record, nil
}

var MockStorageError = errors.New("mock error")

type MockErrorStorage struct{}

func (s *MockErrorStorage) Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error) {
	return DNSResponse{}, MockStorageError
}

func (s *MockErrorStorage) CreateRecord(ctx context.Context, p RecordCreateParameters) (Record, error) {
	return Record{}, MockStorageError
}

type MockStorageOptions struct {
	ReturnErrors bool
}

func NewMockService(o MockStorageOptions) Storage {
	if o.ReturnErrors {
		return &MockErrorStorage{}
	}
	return &MockStorage{
		nextID:  1,
		records: []Record{},
	}
}
