package resolver

import (
	"context"

	"git.houseofkummer.com/lior/home-dns/api/resolver"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type service struct {
	resolver.UnimplementedResolverServer
	handler Resolver
	logger  *zap.Logger
}

func Register(e *grpc.Server, s storage.Storage, l *zap.Logger) {
	resolver.RegisterResolverServer(e, &service{
		handler: s,
		logger:  l,
	})
}

func (s *service) Resolve(ctx context.Context, q *resolver.Question) (*resolver.Response, error) {
	resp, err := s.handler.Resolve(ctx, storage.DNSQuestion{
		Name:  q.Name,
		Qtype: uint16(q.Qtype),
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(
		"DNS request",
		zap.String("name", q.Name),
	)
	return &resolver.Response{
		Answer: resp.Answer,
		Ns:     resp.NS,
		Extra:  resp.Extra,
	}, nil
}
