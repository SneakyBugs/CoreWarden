package storage

import (
	"context"
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
		ID:         s.nextID,
		Zone:       p.Zone,
		RR:         p.RR,
		Comment:    p.Comment,
		CreatedAt:  time.Now(),
		ModifiedOn: time.Now(),
	}
	s.records = append(s.records, record)
	s.nextID++
	return record, nil
}

func (s *MockStorage) ReadRecord(ctx context.Context, id int) (Record, error) {
	for _, r := range s.records {
		if r.ID == id {
			return r, nil
		}
	}
	return Record{}, RecordNotFoundError
}

type MockErrorStorage struct {
	Error error
}

func (s *MockErrorStorage) Resolve(ctx context.Context, q DNSQuestion) (DNSResponse, error) {
	return DNSResponse{}, s.Error
}

func (s *MockErrorStorage) CreateRecord(ctx context.Context, p RecordCreateParameters) (Record, error) {
	return Record{}, s.Error
}

func (s *MockErrorStorage) ReadRecord(ctx context.Context, id int) (Record, error) {
	return Record{}, s.Error
}

type MockStorageOptions struct {
	ReturnError error
}

func NewMockService(o MockStorageOptions) Storage {
	if o.ReturnError != nil {
		return &MockErrorStorage{
			Error: o.ReturnError,
		}
	}
	return &MockStorage{
		nextID:  1,
		records: []Record{},
	}
}
