package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
			return h.replyText(ctx, event.ReplyToken, "請點擊下方選單開始上傳餐點")
		}

		h.cache.SetMealDescription(userID, text)

		mealName := getMealName(meal)

		vars := []template.Variable{
			{Name: "餐點", Value: mealName},
			{Name: "名稱", Value: text},
		}
		msg := template.GetCheckMsg("上傳餐點",
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
		return h.replyText(ctx, event.ReplyToken, "歡迎使用營養師機器人，請依序輸入您的身高(公分)、體重(公斤)、年齡、性別(男/女)以完成註冊\n\n請輸入身高(公分) ex. 175.5")
	}

	if h.cache.GetHeight(userID) == 0 {
		height, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, "身高格式錯誤，請重新輸入 ex. 175.5")
		}
		h.cache.SetHeight(userID, height)
		return h.replyText(ctx, event.ReplyToken, "請輸入體重(公斤) ex. 70.5")
	}

	if h.cache.GetWeight(userID) == 0 {
		w, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, "體重格式錯誤，請重新輸入 ex. 70.5")
		}
		h.cache.SetWeight(userID, w)
		return h.replyText(ctx, event.ReplyToken, "請輸入年齡 ex. 25")
	}

	if h.cache.GetAge(userID) == 0 {
		a, err := strconv.Atoi(text)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, "年齡格式錯誤，請重新輸入 ex. 25")
		}
		h.cache.SetAge(userID, a)
		return h.replyText(ctx, event.ReplyToken, "請輸入目標體重 ex. 65.0")
	}

	if h.cache.GetTargetWeight(userID) == 0 {
		tw, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, "目標體重格式錯誤，請重新輸入 ex. 65.0")
		}
		h.cache.SetTargetWeight(userID, tw)
		return h.replyText(ctx, event.ReplyToken, "請輸入目標時間(月) ex. 3")
	}

	if h.cache.GetTargetTimeframe(userID) == 0 {
		ttf, err := strconv.Atoi(text)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, "目標時間格式錯誤，請重新輸入 ex. 3")
		}
		h.cache.SetTargetTimeframe(userID, ttf)
		return h.replyText(ctx, event.ReplyToken, "請輸入性別(男/女)")
	}

	if h.cache.GetGender(userID) == 0 {
		t := strings.TrimSpace(text)
		var g models.Gender
		switch t {
		case "男":
			g = models.GenderMale
		case "女":
			g = models.GenderFemale
		default:
			return h.replyText(ctx, event.ReplyToken, "性別格式錯誤，請重新輸入(男/女)")
		}
		h.cache.SetGender(userID, int(g))

		// Confirm Msg
		genderStr := "男"
		if g == models.GenderFemale {
			genderStr = "女"
		}

		vars := []template.Variable{
			{Name: "性別", Value: genderStr},
			{Name: "身高", Value: fmt.Sprintf("%.1f 公分", h.cache.GetHeight(userID))},
			{Name: "體重", Value: fmt.Sprintf("%.1f 公斤", h.cache.GetWeight(userID))},
			{Name: "年齡", Value: fmt.Sprintf("%d 歲", h.cache.GetAge(userID))},
			{Name: "目標體重", Value: fmt.Sprintf("%.1f 公斤", h.cache.GetTargetWeight(userID))},
			{Name: "目標時間", Value: fmt.Sprintf("%d 個月", h.cache.GetTargetTimeframe(userID))},
		}
		msg := template.GetCheckMsg("基本資料", vars, []string{"action=check_basic_info&data=y", "action=check_basic_info&data=n"})
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
			return h.replyText(ctx, event.ReplyToken, "目標體重格式錯誤，請重新輸入 ex. 65.0")
		}
		h.cache.SetTargetWeight(userID, tw)
		return h.replyText(ctx, event.ReplyToken, "請輸入目標時間(月) ex. 3")
	}
	if h.cache.GetTargetTimeframe(userID) == 0 {
		ttf, err = strconv.Atoi(text)
		if err != nil {
			return h.replyText(ctx, event.ReplyToken, "目標時間格式錯誤，請重新輸入 ex. 3")
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
		if err := h.aiAPI.AnalyzeBasicInfo(ctx, uid, userID, 0, &gpt.BasicUserInfo{
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
	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf("目標已更新為 %.1f 公斤，預計 %d 個月達成, AI 分析你的目標中...", tw, ttf))
}

func (h *Handler) recordWaterProcess(ctx context.Context, event *linebot.Event, text string) error {
	water, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, "飲水量格式錯誤，請重新輸入 ex. 500")
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

	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf("飲水量已更新 %.1f 毫升", water))
}

func (h *Handler) recordSleepProcess(ctx context.Context, event *linebot.Event, text string) error {
	sleep, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, "睡眠時數格式錯誤，請重新輸入 ex. 8")
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

	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf("睡眠時數已更新 %.1f 小時", sleep))
}

func (h *Handler) recordWeightProcess(ctx context.Context, event *linebot.Event, text string) error {
	w, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return h.replyText(ctx, event.ReplyToken, "體重格式錯誤，請重新輸入 ex. 70.5")
	}
	defer h.cache.SetTextMessageActionType(event.Source.UserID, action.TextMessageActionTypeUnknown)
	if err := h.store.UpdateUserWeight(ctx, event.Source.UserID, w); err != nil {
		h.logger.Error("Error updating weight", zap.Error(err))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	return h.replyText(ctx, event.ReplyToken, fmt.Sprintf("體重已更新 %.1f 公斤", w))
}
