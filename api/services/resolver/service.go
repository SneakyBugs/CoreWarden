package resolver

import (
	"context"

	"git.houseofkummer.com/lior/home-dns/api/resolver"
	"git.houseofkummer.com/lior/home-dns/api/services/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		if status.Convert(err).Code() == codes.NotFound {
			// Avoid logging records not found as errors.
			s.logger.Info("DNS request", zap.String("name", q.Name), zap.Bool("found", false))
			return nil, err
		}
		s.logger.Error(
			"DNS request error",
			zap.Error(err),
		)
		return nil, err
	}
	s.logger.Info(
		"DNS request",
		zap.String("name", q.Name),
		zap.Bool("found", true),
	)
	return &resolver.Response{
		Answer: resp.Answer,
		Ns:     resp.NS,
		Extra:  resp.Extra,
	}, nil
}
