package handler

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/0x726f6f6b6965/go-nutritionst/service/bot/action"
	pushMsg "github.com/0x726f6f6b6965/go-nutritionst/service/push/msg"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

func (h *Handler) HandlePostback(ctx context.Context, event *linebot.Event) error {
	userID := event.Source.UserID
	data := event.Postback.Data

	values, err := url.ParseQuery(data)
	if err != nil {
		return err
	}

	actionType := action.ToPostbackActionType(values.Get("action"))
	datas := values["data"] // Can be multiple

	user, err := h.getUserInfo(ctx, userID)
	if err != nil {
		h.logger.Error("Error getting user info", zap.Error(err))
		return internalErrors.ErrFaiedToGetUser
	}

	if user == nil && actionType != action.ActionTypeCheckBasicInfo {
		return h.addUserProcess(ctx, event, "")
	}

	switch actionType {
	case action.ActionTypeSetMeal:
		return h.setMeal(ctx, userID, event.ReplyToken, datas)
	case action.ActionTypeCheckDescript:
		return h.checkDescript(ctx, userID, event.ReplyToken, datas)
	case action.ActionTypeSetDescription:
		return h.setDescription(ctx, userID, event.ReplyToken)
	case action.ActionTypeCheckBasicInfo:
		if len(datas) > 0 && datas[0] == "y" {
			return h.checkBasicInfo(ctx, userID, event.ReplyToken)
		} else {
			h.removeRegisterProcess(userID)
			// Restart process
			return h.addUserProcess(ctx, event, "")
		}
	case action.ActionTypeDailyReport:
		return h.dailyReport(ctx, userID, user, event.ReplyToken)
	case action.ActionTypeChangeTarget:
		return h.changeTarget(ctx, userID, event.ReplyToken)
	case action.ActionTypeSetWater:
		return h.setWater(ctx, userID, event.ReplyToken)
	case action.ActionTypeSetSleep:
		return h.setSleep(ctx, userID, event.ReplyToken)
	case action.ActionTypeSetWeight:
		return h.setWeight(ctx, userID, event.ReplyToken)
	case action.ActionTypeSetting:
		return h.setting(ctx, event.ReplyToken)
	case action.ActionTypeChangePushMsg:
		return h.changePushMsg(ctx, event.ReplyToken)
	case action.ActionTypeSetPushMsg:
		return h.setPushMsg(ctx, userID, event.ReplyToken, datas)
	default:
		return h.replyText(ctx, event.ReplyToken, "unknown action")
	}
}

func (h *Handler) setMeal(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) > 0 {
		mealInt, err := strconv.Atoi(datas[0])
		if err != nil {
			h.logger.Error("Error parsing meal", zap.Error(err))
			return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
		}
		h.cache.SetMeal(userID, mealInt)
		mealName := getMealName(mealInt)
		return h.replyText(ctx, replyToken, fmt.Sprintf("請輸入%s餐點名稱", mealName))
	}
	return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
}

func (h *Handler) checkDescript(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) > 0 && datas[0] == "y" {
		meal := h.cache.GetMeal(userID)
		desc := h.cache.GetMealDescription(userID)
		vars := []template.Variable{
			{Name: "餐點", Value: getMealName(meal)},
			{Name: "名稱", Value: desc},
		}
		msg := template.GetUploadMsg(vars)
		return h.replyFlex(ctx, replyToken, "upload img", msg)
	} else {
		h.cache.DeleteMeal(userID)
		h.cache.DeleteMealDescription(userID)
		msg := template.GetSetMealMsg()
		return h.replyFlex(ctx, replyToken, "set description", msg)
	}
}

func (h *Handler) setDescription(ctx context.Context, userID string, replyToken string) error {
	h.cache.DeleteMeal(userID)
	h.cache.DeleteMealDescription(userID)
	msg := template.GetSetMealMsg()
	return h.replyFlex(ctx, replyToken, "set description", msg)
}

func (h *Handler) checkBasicInfo(ctx context.Context, userID string, replyToken string) error {
	height := h.cache.GetHeight(userID)
	weight := h.cache.GetWeight(userID)
	targetWeight := h.cache.GetTargetWeight(userID)
	targetTimeframe := h.cache.GetTargetTimeframe(userID)
	age := h.cache.GetAge(userID)
	gender := models.Gender(h.cache.GetGender(userID))

	newUser := &models.User{
		LineID:           userID,
		Height:           height,
		Weight:           weight,
		TargetWeight:     targetWeight,
		TargetTimeframe:  targetTimeframe,
		Age:              age,
		Gender:           gender,
		MaxDailyToken:    h.maxDailyToken,
		MorningMsgSent:   true,
		EveningMsgSent:   true,
		BreakfastMsgSent: true,
		LunchMsgSent:     true,
		DinnerMsgSent:    true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if err := h.store.CreateUser(ctx, newUser); err != nil {
		h.logger.Error("Error creating user", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrRegister.Error())
	}
	// Send request to AI
	uid := uuid.New()
	if err := h.store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeBasicInfo,
		LineID:      userID,
		Status:      models.SendRequestStatusPending,
	}); err != nil {
		h.logger.Error("Error creating send request", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	h.removeRegisterProcess(userID)
	go func() {
		if err := h.aiAPI.AnalyzeBasicInfo(ctx, uid, userID, 0, &gpt.BasicUserInfo{
			UserProfile: newUser.ToProfileString(),
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
	return h.replyText(ctx, replyToken, "註冊成功! AI 分析目標中，請稍後")
}

func (h *Handler) dailyReport(ctx context.Context, userID string, user *models.User, replyToken string) error {

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.AddDate(0, 0, 1)

	q := query.NewQuery()
	q = q.AddFilter(squirrel.And{
		squirrel.Eq{"line_id": userID},
		squirrel.GtOrEq{"created_at": start},
		squirrel.Lt{"created_at": end},
	}).AddSortBy("created_at", false)

	histories, err := h.store.GetMealHistory(ctx, q)
	if err != nil {
		h.logger.Error("Error getting history", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	uid := uuid.New()
	dailyInfo := &gpt.DailyInfo{
		UserProfile: user.ToProfileString(),
		Meta:        today.Format("2006-01-02"),
	}
	meals := make([]models.Meal, 0)
	for _, history := range histories {
		dailyInfo.MealsToday = append(dailyInfo.MealsToday, gpt.MealInfo{
			Meal:                int(history.Meal),
			Description:         history.Description,
			PreviousDescription: history.AIDescription,
			PreviousAdvice:      history.AISuggest,
		})
		meals = append(meals, history.Meal)
	}
	mealsInt := models.GetMealsIntFromMeals(meals)
	if mealsInt == 0 {
		return h.replyText(ctx, replyToken, internalErrors.ErrNoHistory.Error())
	}
	prev, err := h.store.GetMealDaily(ctx, query.NewQuery().
		AddFilter(squirrel.Eq{"line_id": userID}).
		AddFilter(squirrel.Eq{"meals": mealsInt}))
	if err != nil {
		h.logger.Error("Error getting history", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	if len(prev) > 0 {
		daliy := prev[0]
		respMsg := template.GetAIMealDailyMsg(&daliy, 0)
		msg := messaging_api.FlexMessage{
			AltText:  "AI Meal Response",
			Contents: respMsg,
		}
		return h.sendMsg(ctx, userID, uid.String(), msg)
	}

	usedToken, err := h.store.GetUsage(ctx, userID)
	if err != nil {
		h.logger.Error("Error getting usage", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	if time.Since(usedToken.UpdatedAt) > 24*time.Hour {
		usedToken.Usage = 0
	}
	if usedToken.Usage >= h.maxDailyToken {
		return h.replyText(ctx, replyToken, internalErrors.ErrOutOfDailyToken.Error())
	}
	// Send request to AI
	if err := h.store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeDaily,
		LineID:      userID,
		Status:      models.SendRequestStatusPending,
	}); err != nil {
		h.logger.Error("Error creating send request", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}

	go func() {
		if err := h.aiAPI.AnalyzeDailyMeal(ctx, uid, userID, usedToken.Usage, dailyInfo); err != nil {
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
	return h.replyText(ctx, replyToken, "AI 分析中，請稍後")
}

func (h *Handler) removeRegisterProcess(userID string) {
	h.cache.DeleteRegisterProcess(userID)
	h.cache.DeleteHeight(userID)
	h.cache.DeleteWeight(userID)
	h.cache.DeleteTargetWeight(userID)
	h.cache.DeleteTargetTimeframe(userID)
	h.cache.DeleteAge(userID)
	h.cache.DeleteGender(userID)
}

func (h *Handler) setting(ctx context.Context, replyToken string) error {
	msg := template.GetSettingMsg()
	return h.replyFlex(ctx, replyToken, "setting", msg)
}

func (h *Handler) changeTarget(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeSetTarget)
	return h.replyText(ctx, replyToken, "請輸入新的目標體重(公斤)")
}

func (h *Handler) setWater(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeRecordWater)
	return h.replyText(ctx, replyToken, "請輸入今日飲水量")
}

func (h *Handler) setSleep(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeRecordSleep)
	return h.replyText(ctx, replyToken, "請輸入今日睡眠時數")
}

func (h *Handler) setWeight(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeRecordWeight)
	return h.replyText(ctx, replyToken, "請輸入今日體重")
}

func (h *Handler) changePushMsg(ctx context.Context, replyToken string) error {
	msg := template.GetChangePushMsg()
	return h.replyFlex(ctx, replyToken, "change push msg", msg)
}

func (h *Handler) setPushMsg(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) < 1 {
		msg := template.GetChangePushMsg()
		return h.replyFlex(ctx, replyToken, "change push msg", msg)
	}
	typ := pushMsg.GetMsgType(datas[0])
	if typ == pushMsg.MsgTypeUnknown {
		return h.replyText(ctx, replyToken, "Invalid push message type")
	}
	user, err := h.store.GetUserByLineID(ctx, userID)
	if err != nil {
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	var (
		updateVals []storage.UpdateColumn
		result     string
	)

	switch typ {
	case pushMsg.MsgTypeMorning:
		user.MorningMsgSent = !user.MorningMsgSent
		updateVals = append(updateVals, storage.UpdateColumn{
			ColumnName: storage.UserMorningMsgSent,
			Value:      user.MorningMsgSent,
		})
		if user.MorningMsgSent {
			result = "開啟"
		} else {
			result = "關閉"
		}
	case pushMsg.MsgTypeEvening:
		user.EveningMsgSent = !user.EveningMsgSent
		updateVals = append(updateVals, storage.UpdateColumn{
			ColumnName: storage.UserEveningMsgSent,
			Value:      user.EveningMsgSent,
		})
		if user.EveningMsgSent {
			result = "開啟"
		} else {
			result = "關閉"
		}
	}
	if err := h.store.UpdateUser(ctx, userID, updateVals...); err != nil {
		h.logger.Error("Error updating user", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	return h.replyText(ctx, replyToken, fmt.Sprintf("更新成功, 已將%s推播功能%s", typ.ChineseString(), result))
}
