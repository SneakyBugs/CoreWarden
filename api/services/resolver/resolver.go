package resolver

import (
	"context"
	"fmt"
)

type Resolver interface {
	Resolve(ctx context.Context, q Question) (Response, error)
}

type Question struct {
	Name  string
	Qtype uint16
}

type Response struct {
	Answer []string
	NS     []string
	Extra  []string
}

type LocalhostResolver struct{}

func (r LocalhostResolver) Resolve(ctx context.Context, q Question) (Response, error) {
	return Response{
		Answer: []string{fmt.Sprintf("%s 3600 IN A 127.0.0.1", q.Name)},
	}, nil
}
