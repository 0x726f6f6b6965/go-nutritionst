package bot

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Option func(*Service) error

func WithMaxDailyToken(maxDailyToken int64) Option {
	return func(s *Service) error {
		s.maxDailyToken = maxDailyToken
		return nil
	}
}

func WithAnalyzeMealFn(analyzeMealFn func(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, mealInfo *gpt.MealInfoWithImage, s *Service) error) Option {
	return func(s *Service) error {
		s.AnalyzeMealFn = analyzeMealFn
		return nil
	}
}

func WithAnalyzeMealDailyFn(analyzeMealDailyFn func(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, dailyInfo *gpt.DailyInfo, s *Service) error) Option {
	return func(s *Service) error {
		s.AnalyzeMealDailyFn = analyzeMealDailyFn
		return nil
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(s *Service) error {
		s.logger = logger
		return nil
	}
}
