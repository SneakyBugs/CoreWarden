package slog

import (
	"context"
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const name = "slog"

type SLog struct {
	Next   plugin.Handler
	Logger *zap.Logger
}

func (sl SLog) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	ctxWithLogger := context.WithValue(ctx, contextKeyLogger, sl.Logger)
	returnCode, err := plugin.NextOrFailure(sl.Name(), sl.Next, ctxWithLogger, w, r)

	state := request.Request{W: w, Req: r}
	sl.Logger.Info("client request",
		zap.String("type", state.Type()),
		zap.String("name", state.Name()),
		zap.String("class", state.Class()),
		zap.String("remote", state.IP()),
		zap.Int("size", state.Size()),
		zap.String("rcode", dns.RcodeToString[returnCode]),
		// TODO Add trace ID.
	)
	return returnCode, err
}

func (sl SLog) Name() string {
	return name
}

func LoggerFromContext(ctx context.Context) (*zap.Logger, bool) {
	logger, ok := ctx.Value(contextKeyLogger).(*zap.Logger)
	return logger, ok
}

type contextKey string

func (c contextKey) String() string {
	return fmt.Sprintf("slog:%s", string(c))
}

var (
	contextKeyLogger = contextKey("logger")
)
