package resolver

import (
	"context"

	"git.houseofkummer.com/lior/home-dns/api/services/storage"
)

type Resolver interface {
	Resolve(ctx context.Context, q storage.DNSQuestion) (storage.DNSResponse, error)
}
