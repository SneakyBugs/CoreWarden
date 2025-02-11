package resolver

import (
	"context"

	"github.com/sneakybugs/corewarden/api/services/storage"
)

type Resolver interface {
	Resolve(ctx context.Context, q storage.DNSQuestion) (storage.DNSResponse, error)
}
