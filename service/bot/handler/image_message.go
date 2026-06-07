package handler

import (
	"context"
	"io"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"go.uber.org/zap"
)

func (h *Handler) HandleImageMessage(ctx context.Context, event *linebot.Event, message *linebot.ImageMessage) error {
	userID := event.Source.UserID

	user, err := h.getUserInfo(ctx, userID)
	if err != nil {
		h.logger.Error("Error getting user info", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetUser.Error())
	}
	// Register logic if user not found
	if user == nil {
		return h.addUserProcess(ctx, event, "")
	}

	meal := h.cache.GetMeal(userID)
	description := h.cache.GetMealDescription(userID)

	if meal == 0 || description == "" {
		msg := template.GetSetMealMsg()
		return h.replyFlex(ctx, event.ReplyToken, "set description", msg)
	}

	// Get Image Content
	content, err := h.blobClient.GetMessageContent(message.ID)
	if err != nil {
		h.logger.Error("Error getting image content", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetImage.Error())
	}
	defer content.Body.Close()

	imgData, err := io.ReadAll(content.Body)
	if err != nil {
		h.logger.Error("Error reading image content", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}

	h.cache.SetPicture(userID, imgData)
	// Send request to AI
	uid := uuid.New()
	if err := h.store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeMeal,
		LineID:      userID,
		Status:      models.SendRequestStatusPending,
	}); err != nil {
		h.logger.Error("Error creating send request", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}

	mealInfo := &gpt.MealInfoWithImage{
		Image: imgData,
		MealInfo: gpt.MealInfo{
			Meal:        meal,
			Description: description,
			UserProfile: user.ToProfileString(),
		},
	}

	usedToken, err := h.store.GetUsage(ctx, userID)
	if err != nil && err != pgx.ErrNoRows {
		h.logger.Error("Error getting usage", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	if err == pgx.ErrNoRows {
		usedToken = &models.Usage{
			LineID:    userID,
			Usage:     0,
			CreatedAt: time.Now(),
		}
	}
	lastDate, err := usedToken.GetLastUsedDate()
	if err != nil {
		h.logger.Error("Error getting last used date", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	if time.Since(lastDate) > 24*time.Hour {
		usedToken.Usage = 0
	}
	if usedToken.Usage >= h.maxDailyToken {
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrOutOfDailyToken.Error())
	}

	go func() {
		if err := h.aiAPI.AnalyzeMeal(ctx, uid, userID, usedToken, mealInfo); err != nil {
			h.logger.Error("AnalyzeMeal error", zap.Error(err))
			sendErr := h.store.UpdateSendRequest(ctx, uid.String(), storage.UpdateColumn{
				ColumnName: storage.SendRequestStatus,
				Value:      models.SendRequestStatusFailed,
			}, storage.UpdateColumn{
				ColumnName: storage.SendRequestFailReason,
				Value:      err.Error(),
			})
			if sendErr != nil {
				h.logger.Error("UpdateSendRequest error", zap.Error(sendErr))
			}
		}
	}()
	return h.replyText(ctx, event.ReplyToken, template.DescriptionMsgAIAnalyze.String())
}
