package resolver

import (
	"context"

	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ServerError = status.Error(
	codes.Internal,
	"internal server error",
)
var RecordNotFoundError = status.Error(
	codes.NotFound,
	"record not found",
)

type Resolver interface {
	Resolve(ctx context.Context, q storage.DNSQuestion) (storage.DNSResponse, error)
}
