package bot

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/cache"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/ai"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/handler"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

const (
	DefaultMaxDailyToken int64 = 120000
)

type Service struct {
	lineClient    *messaging_api.MessagingApiAPI
	blobClient    *messaging_api.MessagingApiBlobAPI
	store         *storage.Postgres
	cache         *cache.UserContext
	handler       handler.HandlerInterface
	logger        *zap.Logger
	maxDailyToken int64
}

func NewService(channelToken string, store *storage.Postgres, gpt gpt.NutritionAPI, opts ...Option) (*Service, error) {
	client, err := messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		return nil, err
	}
	blobClient, err := messaging_api.NewMessagingApiBlobAPI(channelToken)
	if err != nil {
		return nil, err
	}
	s := &Service{
		lineClient: client,
		blobClient: blobClient,
		store:      store,
		cache:      cache.NewUserContext(),
		// can be override by option
		maxDailyToken: DefaultMaxDailyToken,
		logger:        zap.NewNop(),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	aiAPI := ai.NewService(s.store, s.cache, gpt, s.sendMsg, s.logger)
	handler := handler.NewHandler(s.store,
		s.cache,
		s.blobClient,
		aiAPI,
		s.replyText,
		s.replyFlex,
		s.sendMsg,
		s.maxDailyToken,
		s.logger)
	s.handler = handler
	s.logger.Info("Service initialized", zap.Int64("maxDailyToken", s.maxDailyToken))
	return s, nil
}

func (s *Service) sendMsg(_ context.Context, userID string, xLineRetryKey string, msg messaging_api.MessageInterface) error {
	_, err := s.lineClient.PushMessage(
		&messaging_api.PushMessageRequest{
			To: userID,
			Messages: []messaging_api.MessageInterface{
				msg,
			},
		},
		xLineRetryKey,
	)
	return err
}

func (s *Service) replyFlex(_ context.Context, replyToken, altText string, container messaging_api.FlexContainerInterface) error {
	_, err := s.lineClient.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.FlexMessage{
					AltText:  altText,
					Contents: container,
				},
			},
		},
	)
	return err
}

func (s *Service) replyText(_ context.Context, replyToken, text string) error {
	_, err := s.lineClient.ReplyMessage(
		&messaging_api.ReplyMessageRequest{
			ReplyToken: replyToken,
			Messages: []messaging_api.MessageInterface{
				&messaging_api.TextMessage{
					Text: text,
				},
			},
		},
	)
	return err
}

func (s *Service) HandleEvent(ctx context.Context, event *linebot.Event) error {
	switch event.Type {
	case linebot.EventTypeMessage:
		switch message := event.Message.(type) {
		case *linebot.TextMessage:
			return s.handler.HandleTextMessage(ctx, event, message)
		case *linebot.ImageMessage:
			return s.handler.HandleImageMessage(ctx, event, message)
		}
	case linebot.EventTypePostback:
		return s.handler.HandlePostback(ctx, event)
	}
	return nil
}
