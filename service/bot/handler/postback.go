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
	case action.ActionTypeTargetSuggestion:
		return h.confirmTarget(ctx, userID, event.ReplyToken, datas)
	default:
		h.logger.Error("Unknown action type", zap.String("action_type", values.Get("action")))
		return h.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
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
		return h.replyText(ctx, replyToken, fmt.Sprintf(template.DescriptionMsgAskSendMealName.String(), mealName))
	}
	return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
}

func (h *Handler) checkDescript(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) > 0 && datas[0] == "y" {
		meal := h.cache.GetMeal(userID)
		desc := h.cache.GetMealDescription(userID)
		vars := []template.Variable{
			{Name: template.DescriptionMsgMeal.String(), Value: getMealName(meal)},
			{Name: template.DescriptionMsgMealName.String(), Value: desc},
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
	name := ""
	profile, err := h.lineClient.GetProfile(userID)
	if err != nil {
		h.logger.Error("Error getting user profile", zap.Error(err))
	} else {
		name = profile.DisplayName
	}
	newUser := &models.User{
		LineID:           userID,
		Name:             name,
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
		usedToken := &models.Usage{
			LineID:    userID,
			Usage:     0,
			CreatedAt: time.Now(),
		}
		if err := h.aiAPI.AnalyzeBasicInfo(ctx, uid, userID, usedToken, &gpt.BasicUserInfo{
			UserProfile:     newUser.ToProfileString(),
			TargetWeight:    newUser.TargetWeight,
			TargetTimeframe: newUser.TargetTimeframe,
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
	return h.replyText(ctx, replyToken, fmt.Sprintf("%s%s", template.DescriptionMsgSignUpSuccess.String(), template.DescriptionMsgAIAnalyze.String()))
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
	newQ := query.NewQuery().AddFilter(squirrel.Eq{"line_id": userID})
	if models.GetMealsIntFromMeals(meals, newQ) == 0 {
		return h.replyText(ctx, replyToken, internalErrors.ErrNoHistory.Error())
	}
	prev, err := h.store.GetMealDaily(ctx, newQ)
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
	lastDate, err := usedToken.GetLastUsedDate()
	if err != nil {
		h.logger.Error("Error getting last used date", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	if time.Since(lastDate) > 24*time.Hour {
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
		if err := h.aiAPI.AnalyzeDailyMeal(ctx, uid, userID, usedToken, dailyInfo); err != nil {
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
	return h.replyText(ctx, replyToken, template.DescriptionMsgAIAnalyze.String())
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
	return h.replyText(ctx, replyToken, template.DescriptionMsgAskSetNewTargetWeight.String())
}

func (h *Handler) setWater(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeRecordWater)
	return h.replyText(ctx, replyToken, template.DescriptionMsgAskSendWater.String())
}

func (h *Handler) setSleep(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeRecordSleep)
	return h.replyText(ctx, replyToken, template.DescriptionMsgAskSendSleep.String())
}

func (h *Handler) setWeight(ctx context.Context, userID string, replyToken string) error {
	h.cache.SetTextMessageActionType(userID, action.TextMessageActionTypeRecordWeight)
	return h.replyText(ctx, replyToken, template.DescriptionMsgAskSendWeight.String())
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
		h.logger.Error("Invalid push message type", zap.String("user_id", userID), zap.Strings("datas", datas))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
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
			result = template.DescriptionMsgOpen.String()
		} else {
			result = template.DescriptionMsgClose.String()
		}
	case pushMsg.MsgTypeEvening:
		user.EveningMsgSent = !user.EveningMsgSent
		updateVals = append(updateVals, storage.UpdateColumn{
			ColumnName: storage.UserEveningMsgSent,
			Value:      user.EveningMsgSent,
		})
		if user.EveningMsgSent {
			result = template.DescriptionMsgOpen.String()
		} else {
			result = template.DescriptionMsgClose.String()
		}
	case pushMsg.MsgTypeBreakfast:
		user.BreakfastMsgSent = !user.BreakfastMsgSent
		updateVals = append(updateVals, storage.UpdateColumn{
			ColumnName: storage.UserBreakfastMsgSent,
			Value:      user.BreakfastMsgSent,
		})
		if user.BreakfastMsgSent {
			result = template.DescriptionMsgOpen.String()
		} else {
			result = template.DescriptionMsgClose.String()
		}
	case pushMsg.MsgTypeLunch:
		user.LunchMsgSent = !user.LunchMsgSent
		updateVals = append(updateVals, storage.UpdateColumn{
			ColumnName: storage.UserLunchMsgSent,
			Value:      user.LunchMsgSent,
		})
		if user.LunchMsgSent {
			result = template.DescriptionMsgOpen.String()
		} else {
			result = template.DescriptionMsgClose.String()
		}
	case pushMsg.MsgTypeDinner:
		user.DinnerMsgSent = !user.DinnerMsgSent
		updateVals = append(updateVals, storage.UpdateColumn{
			ColumnName: storage.UserDinnerMsgSent,
			Value:      user.DinnerMsgSent,
		})
		if user.DinnerMsgSent {
			result = template.DescriptionMsgOpen.String()
		} else {
			result = template.DescriptionMsgClose.String()
		}
	}
	if err := h.store.UpdateUser(ctx, userID, updateVals...); err != nil {
		h.logger.Error("Error updating user", zap.Error(err))
		return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	return h.replyText(ctx, replyToken, fmt.Sprintf(template.DescriptionMsgUpdateAlermSuccess.String(), typ.ChineseString(), result))
}

func (h *Handler) confirmTarget(ctx context.Context, userID string, replyToken string, datas []string) error {
	if datas[0] == action.TargetSuggestionActionTypeKeepPlan.String() {
		profile, err := h.lineClient.GetProfile(userID)
		if err != nil {
			h.logger.Error("Error getting profile", zap.Error(err))
			return h.replyText(ctx, replyToken,
				fmt.Sprintf("Hi,\n%s",
					template.DescriptionMsgStartUse.String(),
				),
			)
		}
		return h.replyText(ctx, replyToken,
			fmt.Sprintf("Hi %s,\n%s",
				profile.DisplayName,
				template.DescriptionMsgStartUse.String(),
			),
		)
	}
	if datas[0] == action.TargetSuggestionActionTypeChangePlan.String() {
		if len(datas) < 3 {
			h.logger.Error("Error parsing target suggestion", zap.Strings("datas", datas))
			return h.replyText(ctx, replyToken, internalErrors.ErrInvalidTargetSuggestion.Error())
		}

		timeFrame, err := strconv.Atoi(datas[1])
		if err != nil {
			h.logger.Error("Error parsing target suggestion", zap.Error(err))
			return h.replyText(ctx, replyToken, internalErrors.ErrInvalidTargetSuggestion.Error())
		}
		targetWeight, err := strconv.ParseFloat(datas[2], 64)
		if err != nil {
			h.logger.Error("Error parsing target suggestion", zap.Error(err))
			return h.replyText(ctx, replyToken, internalErrors.ErrInvalidTargetSuggestion.Error())
		}
		updateVals := []storage.UpdateColumn{
			{
				ColumnName: storage.UserTargetTimeframe,
				Value:      timeFrame,
			},
			{
				ColumnName: storage.UserTargetWeight,
				Value:      targetWeight,
			},
		}
		if err := h.store.UpdateUser(ctx, userID, updateVals...); err != nil {
			h.logger.Error("Error updating user", zap.Error(err))
			return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
		}
		profile, err := h.lineClient.GetProfile(userID)
		if err != nil {
			h.logger.Error("Error getting profile", zap.Error(err))
			return h.replyText(ctx, replyToken,
				fmt.Sprintf("Hi,\n%s",
					template.DescriptionMsgStartUse.String(),
				),
			)
		}
		return h.replyText(ctx, replyToken,
			fmt.Sprintf("Hi %s,\n%s",
				profile.DisplayName,
				template.DescriptionMsgStartUse.String(),
			),
		)
	}
	h.logger.Error("Invalid target suggestion", zap.Strings("datas", datas))
	return h.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
}
