package resolver

import (
	"context"
	"log"

	"git.houseofkummer.com/lior/home-dns/api/resolver"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"google.golang.org/grpc"
)

type service struct {
	resolver.UnimplementedResolverServer
	handler Resolver
}

func Register(e *grpc.Server, s storage.Storage) {
	resolver.RegisterResolverServer(e, &service{
		handler: s,
	})
}

func (s *service) Resolve(ctx context.Context, q *resolver.Question) (*resolver.Response, error) {
	log.Printf("Received query for %v\n", q.Name)
	resp, err := s.handler.Resolve(ctx, storage.DNSQuestion{
		Name:  q.Name,
		Qtype: uint16(q.Qtype),
	})
	if err != nil {
		return nil, err
	}
	return &resolver.Response{
		Answer: resp.Answer,
		Ns:     resp.NS,
		Extra:  resp.Extra,
	}, nil
}
