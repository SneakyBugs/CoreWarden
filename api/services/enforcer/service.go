package enforcer

import "go.uber.org/zap"

type Service struct {
	enforcer Enforcer
	logger   *zap.Logger
}

func NewService(e Enforcer, l *zap.Logger) Service {
	return Service{
		enforcer: e,
		logger:   l,
	}
}

func (s *Service) NewRequestEnforcer(object string) RequestEnforcer {
	return NewRequestEnforcer(s.enforcer, object)
}
