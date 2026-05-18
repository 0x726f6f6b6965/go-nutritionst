package ai

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/cache"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

var _ NutritionAPI = (*Service)(nil)

type NutritionAPI interface {
	AnalyzeMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken *models.Usage, mealInfo *gpt.MealInfoWithImage) error
	AnalyzeDailyMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken *models.Usage, dailyInfo *gpt.DailyInfo) error
	AnalyzeBasicInfo(ctx context.Context, uid uuid.UUID, userID string, usedToken *models.Usage, basicInfo *gpt.BasicUserInfo) error
}

type Service struct {
	store   *storage.Postgres
	cache   *cache.UserContext
	gpt     gpt.NutritionAPI
	sendMsg func(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error
	logger  *zap.Logger
}

func NewService(store *storage.Postgres,
	cache *cache.UserContext,
	gpt gpt.NutritionAPI,
	sendMsg func(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error,
	logger *zap.Logger) *Service {
	return &Service{
		store:   store,
		cache:   cache,
		gpt:     gpt,
		sendMsg: sendMsg,
		logger:  logger,
	}
}
