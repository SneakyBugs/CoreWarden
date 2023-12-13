package resolver

import (
	"context"
	"errors"

	"git.houseofkummer.com/lior/home-dns/api/services/storage"
)

var RecordNotFoundError = errors.New("record not found")

type Resolver interface {
	Resolve(ctx context.Context, q storage.DNSQuestion) (storage.DNSResponse, error)
}
