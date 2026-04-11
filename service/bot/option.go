package bot

import (
	"go.uber.org/zap"
)

type Option func(*Service) error

func WithMaxDailyToken(maxDailyToken int64) Option {
	return func(s *Service) error {
		s.maxDailyToken = maxDailyToken
		return nil
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(s *Service) error {
		s.logger = logger
		return nil
	}
}
