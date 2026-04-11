package handler

import (
	"context"
	"strings"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/cache"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/ai"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

var _ HandlerInterface = (*Handler)(nil)

type HandlerInterface interface {
	HandlePostback(ctx context.Context, event *linebot.Event) error
	HandleTextMessage(ctx context.Context, event *linebot.Event, message *linebot.TextMessage) error
	HandleImageMessage(ctx context.Context, event *linebot.Event, message *linebot.ImageMessage) error
}

type Handler struct {
	store         *storage.Postgres
	cache         *cache.UserContext
	blobClient    *messaging_api.MessagingApiBlobAPI
	aiAPI         ai.NutritionAPI
	replyText     func(_ context.Context, replyToken, text string) error
	replyFlex     func(_ context.Context, replyToken, altText string, container messaging_api.FlexContainerInterface) error
	sendMsg       func(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error
	maxDailyToken int64
	logger        *zap.Logger
}

func NewHandler(store *storage.Postgres,
	cache *cache.UserContext,
	blobClient *messaging_api.MessagingApiBlobAPI,
	aiAPI ai.NutritionAPI,
	replyText func(_ context.Context, replyToken, text string) error,
	replyFlex func(_ context.Context, replyToken, altText string, container messaging_api.FlexContainerInterface) error,
	sendMsg func(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error,
	maxDailyToken int64,
	logger *zap.Logger) *Handler {
	return &Handler{
		store:         store,
		cache:         cache,
		blobClient:    blobClient,
		replyText:     replyText,
		replyFlex:     replyFlex,
		sendMsg:       sendMsg,
		aiAPI:         aiAPI,
		maxDailyToken: maxDailyToken,
		logger:        logger,
	}
}

func (h *Handler) getUserInfo(ctx context.Context, userID string) (*models.User, error) {
	user, err := h.store.GetUserByLineID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") { // pgx v5 returns ErrNoRows equivalent
			return nil, nil // Not found
		}
		// squirrel/pgx might return different error structure, but usually pgx.ErrNoRows
		// Check if it's actually no rows
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func getMealName(m int) string {
	switch m {
	case 1:
		return "早餐"
	case 2:
		return "午餐"
	case 3:
		return "晚餐"
	case 4:
		return "點心"
	default:
		return "未知"
	}
}
