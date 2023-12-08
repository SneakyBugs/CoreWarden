package resolver

import (
	"context"
	"fmt"

	"git.houseofkummer.com/lior/home-dns/api/services/storage"
)

type Resolver interface {
	Resolve(ctx context.Context, q storage.DNSQuestion) (storage.DNSResponse, error)
}

type LocalhostResolver struct{}

func (r LocalhostResolver) Resolve(ctx context.Context, q storage.DNSQuestion) (storage.DNSResponse, error) {
	return storage.DNSResponse{
		Answer: []string{fmt.Sprintf("%s 3600 IN A 127.0.0.1", q.Name)},
	}, nil
}
