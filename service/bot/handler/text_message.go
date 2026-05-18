package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/timezone"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/action"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"go.uber.org/zap"
)

func (h *Handler) HandleTextMessage(ctx context.Context, event *linebot.Event, message *linebot.TextMessage) error {
	userID := event.Source.UserID
	text := message.Text

	user, err := h.getUserInfo(ctx, userID)
	if err != nil {
		h.logger.Error("Error getting user info", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetUser.Error())
	}

	// Register logic
	if user == nil || h.cache.GetRegisterProcess(userID) {
		return h.addUserProcess(ctx, event, text)
	}

	// check the text message action type
	textMessageActionType := h.cache.GetTextMessageActionType(userID)
	switch textMessageActionType {
	case action.TextMessageActionTypeSetTarget:
		return h.changeTargetWeightProcess(ctx, event, text)
	case action.TextMessageActionTypeRecordWater:
		return h.recordWaterProcess(ctx, event, text)
	case action.TextMessageActionTypeRecordSleep:
		return h.recordSleepProcess(ctx, event, text)
	case action.TextMessageActionTypeRecordWeight:
		return h.recordWeightProcess(ctx, event, text)
	default:
		// Registered user: Meal Logic
		meal := h.cache.GetMeal(userID)
		if meal == 0 {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgUploadMealIntro.String())
		}

		h.cache.SetMealDescription(userID, text)

		mealName := getMealName(meal)

		vars := []template.Variable{
			{Name: DescriptionMsgMeal.String(), Value: mealName},
			{Name: DescriptionMsgMealName.String(), Value: text},
		}
		msg := template.GetCheckMsg(DescriptionMsgUploadMeal.String(),
			vars,
			[]string{fmt.Sprintf("action=%s&data=y", action.ActionTypeCheckDescript.String()),
				fmt.Sprintf("action=%s&data=n", action.ActionTypeCheckDescript.String())})

		return h.replyFlex(ctx, event.ReplyToken, "upload img", msg)
	}
}

func (h *Handler) addUserProcess(ctx context.Context, event *linebot.Event, text string) error {
	userID := event.Source.UserID

	if !h.cache.GetRegisterProcess(userID) {
		h.cache.SetRegisterProcess(userID, true)
		return h.replyText(ctx, event.ReplyToken, fmt.Sprintf("%s\n\n%s", DescriptionMsgWelcomeSignUp.String(), DescriptionMsgAskHeight.String()))
	}

	if h.cache.GetHeight(userID) == 0 {
		height, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgHeightFormatError.String())
		}
		h.cache.SetHeight(userID, height)
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgAskWeight.String())
	}

	if h.cache.GetWeight(userID) == 0 {
		w, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgWeightFormatError.String())
		}
		h.cache.SetWeight(userID, w)
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgAskAge.String())
	}

	if h.cache.GetAge(userID) == 0 {
		a, err := strconv.Atoi(text)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgAgeFormatError.String())
		}
		h.cache.SetAge(userID, a)
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgAskSetTargetWeight.String())
	}

	if h.cache.GetTargetWeight(userID) == 0 {
		tw, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgTargetWeightFormatError.String())
		}
		h.cache.SetTargetWeight(userID, tw)
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgAskSetTargetTime.String())
	}

	if h.cache.GetTargetTimeframe(userID) == 0 {
		ttf, err := strconv.Atoi(text)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgTargetTimeFormatError.String())
		}
		h.cache.SetTargetTimeframe(userID, ttf)
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgAskSetGender.String())
	}

	if h.cache.GetGender(userID) == 0 {
		t := strings.TrimSpace(text)
		var g models.Gender
		switch t {
		case DescriptionMsgMale.String():
			g = models.GenderMale
		case DescriptionMsgFemale.String():
			g = models.GenderFemale
		default:
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgGenderFormatError.String())
		}
		h.cache.SetGender(userID, int(g))

		// Confirm Msg
		genderStr := DescriptionMsgMale.String()
		if g == models.GenderFemale {
			genderStr = DescriptionMsgFemale.String()
		}

		vars := []template.Variable{
			{Name: DescriptionMsgGender.String(), Value: genderStr},
			{Name: DescriptionMsgHeight.String(), Value: fmt.Sprintf("%.1f 公分", h.cache.GetHeight(userID))},
			{Name: DescriptionMsgWeight.String(), Value: fmt.Sprintf("%.1f 公斤", h.cache.GetWeight(userID))},
			{Name: DescriptionMsgAge.String(), Value: fmt.Sprintf("%d 歲", h.cache.GetAge(userID))},
			{Name: DescriptionMsgTargetWeight.String(), Value: fmt.Sprintf("%.1f 公斤", h.cache.GetTargetWeight(userID))},
			{Name: DescriptionMsgTargetTime.String(), Value: fmt.Sprintf("%d 個月", h.cache.GetTargetTimeframe(userID))},
		}
		msg := template.GetCheckMsg(DescriptionMsgBasicInfo.String(), vars, []string{"action=check_basic_info&data=y", "action=check_basic_info&data=n"})
		return h.replyFlex(ctx, event.ReplyToken, "basic info", msg)
	}

	return nil
}

func (h *Handler) changeTargetWeightProcess(ctx context.Context, event *linebot.Event, text string) error {
	userID := event.Source.UserID
	var (
		tw  float64
		ttf int
		err error
	)
	if h.cache.GetTargetWeight(userID) == 0 {
		tw, err = strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgTargetWeightFormatError.String())
		}
		h.cache.SetTargetWeight(userID, tw)
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgAskSetTargetTime.String())
	}
	if h.cache.GetTargetTimeframe(userID) == 0 {
		ttf, err = strconv.Atoi(text)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, DescriptionMsgTargetTimeFormatError.String())
		}
	}
	tw = h.cache.GetTargetWeight(userID)

	if err := h.store.UpdateUserTarget(ctx, event.Source.UserID, tw, ttf); err != nil {
		h.logger.Error("Error updating target weight", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	h.cache.DeleteTargetWeight(userID)
	h.cache.DeleteTargetTimeframe(userID)
	h.cache.SetTextMessageActionType(event.Source.UserID, action.TextMessageActionTypeUnknown)
	uid := uuid.New()
	user, err := h.store.GetUserByLineID(ctx, userID)
	if err != nil {
		h.logger.Error("Error getting user", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	if err := h.store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeBasicInfo,
		LineID:      userID,
		Status:      models.SendRequestStatusPending,
	}); err != nil {
		h.logger.Error("Error creating send request", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	h.cache.DeleteTargetWeight(userID)
	h.cache.DeleteTargetTimeframe(userID)
	h.cache.SetTextMessageActionType(event.Source.UserID, action.TextMessageActionTypeUnknown)
	go func() {
		usedToken := &models.Usage{
			LineID:    userID,
			Usage:     0,
			CreatedAt: time.Now(),
		}
		if err := h.aiAPI.AnalyzeBasicInfo(ctx, uid, userID, usedToken, &gpt.BasicUserInfo{
			UserProfile:     user.ToProfileString(),
			TargetWeight:    user.TargetWeight,
			TargetTimeframe: user.TargetTimeframe,
		}); err != nil {
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
	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf("%s%s", fmt.Sprintf(DescriptionMsgTargetWeightUpdate.String(), tw, ttf), DescriptionMsgAIAnalyze.String()))
}

func (h *Handler) recordWaterProcess(ctx context.Context, event *linebot.Event, text string) error {
	water, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgWaterFormatError.String())
	}
	defer h.cache.SetTextMessageActionType(event.Source.UserID, action.TextMessageActionTypeUnknown)
	date := timezone.GetTaipeiDate()
	q := query.NewQuery().AddFilter(squirrel.Eq{"line_id": event.Source.UserID}).AddFilter(squirrel.Eq{"date": date})
	records, err := h.store.GetDailyRecord(ctx, q)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	if len(records) > 0 {
		if err := h.store.UpdateDailyRecord(ctx, event.Source.UserID, date, storage.UpdateColumn{
			ColumnName: storage.DailyRecordTotalWaterMl,
			Value:      &water,
		}); err != nil {
			h.logger.Error("Error updating water", zap.Error(err))
			return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
		}
	} else {
		info := models.DailyRecord{
			RequestID:    uuid.New(),
			LineID:       event.Source.UserID,
			Date:         date,
			TotalWaterMl: &water,
		}
		if err := h.store.CreateDailyRecord(ctx, &info); err != nil {
			h.logger.Error("Error creating water", zap.Error(err))
			return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
		}
	}

	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf(DescriptionMsgWaterUpdate.String(), water))
}

func (h *Handler) recordSleepProcess(ctx context.Context, event *linebot.Event, text string) error {
	sleep, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgSleepFormatError.String())
	}
	defer h.cache.SetTextMessageActionType(event.Source.UserID, action.TextMessageActionTypeUnknown)
	date := timezone.GetTaipeiDate()
	q := query.NewQuery().AddFilter(squirrel.Eq{"line_id": event.Source.UserID}).AddFilter(squirrel.Eq{"date": date})
	records, err := h.store.GetDailyRecord(ctx, q)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	if len(records) > 0 {
		if err := h.store.UpdateDailyRecord(ctx, event.Source.UserID, date, storage.UpdateColumn{
			ColumnName: storage.DailyRecordTotalSleepHour,
			Value:      &sleep,
		}); err != nil {
			h.logger.Error("Error updating sleep", zap.Error(err))
			return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
		}
	} else {
		info := models.DailyRecord{
			RequestID:      uuid.New(),
			LineID:         event.Source.UserID,
			Date:           date,
			TotalSleepHour: &sleep,
		}
		if err := h.store.CreateDailyRecord(ctx, &info); err != nil {
			h.logger.Error("Error creating sleep", zap.Error(err))
			return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
		}
	}

	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf(DescriptionMsgSleepUpdate.String(), sleep))
}

func (h *Handler) recordWeightProcess(ctx context.Context, event *linebot.Event, text string) error {
	w, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, DescriptionMsgWeightFormatError.String())
	}
	defer h.cache.SetTextMessageActionType(event.Source.UserID, action.TextMessageActionTypeUnknown)
	if err := h.store.UpdateUserWeight(ctx, event.Source.UserID, w); err != nil {
		h.logger.Error("Error updating weight", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf(DescriptionMsgWeightUpdate.String(), w))
}
