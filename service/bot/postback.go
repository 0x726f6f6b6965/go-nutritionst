package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

func (s *Service) setMeal(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) > 0 {
		mealInt, err := strconv.Atoi(datas[0])
		if err != nil {
			s.logger.Error("Error parsing meal", zap.Error(err))
			return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
		}
		s.cache.SetMeal(userID, mealInt)
		mealName := getMealName(mealInt)
		return s.replyText(ctx, replyToken, fmt.Sprintf("請輸入%s餐點名稱", mealName))
	}
	return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
}

func (s *Service) checkDescript(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) > 0 && datas[0] == "y" {
		meal := s.cache.GetMeal(userID)
		desc := s.cache.GetMealDescription(userID)
		vars := []template.Variable{
			{Name: "餐點", Value: getMealName(meal)},
			{Name: "名稱", Value: desc},
		}
		msg := template.GetUploadMsg(vars)
		return s.replyFlex(ctx, replyToken, "upload img", msg)
	} else {
		s.cache.DeleteMeal(userID)
		s.cache.DeleteMealDescription(userID)
		msg := template.GetSetMealMsg()
		return s.replyFlex(ctx, replyToken, "set description", msg)
	}
}

func (s *Service) setDescription(ctx context.Context, userID string, replyToken string) error {
	s.cache.DeleteMeal(userID)
	s.cache.DeleteMealDescription(userID)
	msg := template.GetSetMealMsg()
	return s.replyFlex(ctx, replyToken, "set description", msg)
}

func (s *Service) checkBasicInfo(ctx context.Context, userID string, replyToken string) error {
	height := s.cache.GetHeight(userID)
	weight := s.cache.GetWeight(userID)
	age := s.cache.GetAge(userID)
	gender := models.Gender(s.cache.GetGender(userID))

	newUser := &models.User{
		LineID:    userID,
		Height:    height,
		Weight:    weight,
		Age:       age,
		Gender:    gender,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.store.CreateUser(ctx, newUser); err != nil {
		s.logger.Error("Error creating user", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrRegister.Error())
	}
	s.removeRegisterProcess(userID)
	return s.replyText(ctx, replyToken, "註冊成功! 歡迎使用營養師機器人")
}

func (s *Service) dailyReport(ctx context.Context, userID string, user *models.User, replyToken string) error {

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.AddDate(0, 0, 1)

	q := query.NewQuery()
	q = q.AddFilter(squirrel.And{
		squirrel.Eq{"line_id": userID},
		squirrel.GtOrEq{"created_at": start},
		squirrel.Lt{"created_at": end},
	}).AddSortBy("created_at", false)

	histories, err := s.store.GetMealHistory(ctx, q)
	if err != nil {
		s.logger.Error("Error getting history", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
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
		return s.replyText(ctx, replyToken, internalErrors.ErrNoHistory.Error())
	}
	prev, err := s.store.GetMealDaily(ctx, query.NewQuery().
		AddFilter(squirrel.Eq{"line_id": userID}).
		AddFilter(squirrel.Eq{"meals": mealsInt}))
	if err != nil {
		s.logger.Error("Error getting history", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	if len(prev) > 0 {
		daliy := prev[0]
		respMsg := template.GetAIMealDailyMsg(&daliy, 0)
		msg := messaging_api.FlexMessage{
			AltText:  "AI Meal Response",
			Contents: respMsg,
		}
		return s.sendMsg(ctx, userID, uid.String(), msg)
	}

	usedToken, err := s.store.GetUsage(ctx, userID)
	if err != nil {
		s.logger.Error("Error getting usage", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	if time.Since(usedToken.UpdatedAt) > 24*time.Hour {
		usedToken.Usage = 0
	}
	if usedToken.Usage >= s.maxDailyToken {
		return s.replyText(ctx, replyToken, internalErrors.ErrOutOfDailyToken.Error())
	}
	// Send request to AI
	if err := s.store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeDaily,
		Status:      models.SendRequestStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}); err != nil {
		s.logger.Error("Error creating send request", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}

	go func() {
		if err := s.AnalyzeMealDailyFn(ctx, uid, userID, usedToken.Usage, dailyInfo, s); err != nil {
			s.logger.Error("AnalyzeMeal error", zap.Error(err))
			sendErr := s.store.UpdateSendRequest(ctx, &models.SendRequest{
				RequestID: uid.String(),
				Status:    models.SendRequestStatusFailed,
				Error:     err.Error(),
				UpdatedAt: time.Now(),
			})
			if sendErr != nil {
				s.logger.Error("UpdateSendRequest error", zap.Error(sendErr))
			}
		}
	}()
	return s.replyText(ctx, replyToken, "AI 分析中，請稍後")
}

func (s *Service) joinUs(ctx context.Context, replyToken string) error {
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	msg := template.GetMonthMsg(start, start, "set_report_start")
	return s.replyFlex(ctx, replyToken, "set report start", msg)
}

func (s *Service) getMonth(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) < 2 {
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	dateStr := datas[0]
	nextAct := datas[1]

	giveDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}
	if nextAct != ActionTypeSetReportStart.String() && nextAct != ActionTypeSetReportEnd.String() {
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	altText := ""
	if nextAct == ActionTypeSetReportStart.String() {
		altText = "set report start"
	} else {
		altText = "set report end"
	}

	startDateStr := s.cache.GetStartReport(userID)
	var startDate time.Time
	if startDateStr == "" {
		startDate = time.Now()
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			s.logger.Error("Error parsing date", zap.Error(err))
			return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
		}
	}

	msg := template.GetMonthMsg(giveDate, startDate, nextAct)
	return s.replyFlex(ctx, replyToken, altText, msg)
}

func (s *Service) setReportStart(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) < 1 {
		s.logger.Error("Error parsing date", zap.Strings("datas", datas))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	dateStr := datas[0]
	s.cache.SetStartReport(userID, dateStr)

	startDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		s.logger.Error("Error parsing date", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	msg := template.GetMonthMsg(startDate, startDate, "set_report_end")
	return s.replyFlex(ctx, replyToken, "set report end", msg)
}

func (s *Service) setReportEnd(ctx context.Context, userID string, replyToken string, datas []string) error {
	if len(datas) < 1 {
		s.logger.Error("Error parsing date", zap.Strings("datas", datas))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}
	dateStr := datas[0]
	startDateStr := s.cache.GetStartReport(userID)
	if startDateStr == "" {
		// set the report start first
		return s.replyFlex(ctx, replyToken, "set report start", template.GetMonthMsg(time.Now(), time.Now(), "set_report_start"))
	}
	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", dateStr)
	if endDate.Before(startDate) {
		return s.replyText(ctx, replyToken, internalErrors.ErrEndDateBeforeStartDate.Error())
	}
	q := query.NewQuery()
	q = q.AddFilter(squirrel.And{
		squirrel.Eq{"line_id": userID},
		squirrel.GtOrEq{"created_at": startDate},
		squirrel.Lt{"created_at": endDate.AddDate(0, 0, 1)},
	}).AddSortBy("created_at", false)

	histories, err := s.store.GetMealHistory(ctx, q)
	if err != nil {
		s.logger.Error("Error getting history", zap.Error(err))
		return s.replyText(ctx, replyToken, internalErrors.ErrInternal.Error())
	}

	// Simple JSON dump for now
	b, _ := json.MarshalIndent(histories, "", "  ")
	return s.replyText(ctx, replyToken, string(b))
}

func (s *Service) handlePostback(ctx context.Context, event *linebot.Event) error {
	userID := event.Source.UserID
	data := event.Postback.Data

	values, err := url.ParseQuery(data)
	if err != nil {
		return err
	}

	actionType := ToActionType(values.Get("action"))
	datas := values["data"] // Can be multiple

	user, err := s.getUserInfo(ctx, userID)
	if err != nil {
		s.logger.Error("Error getting user info", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetUser.Error())
	}

	if user == nil && actionType != ActionTypeCheckBasicInfo {
		return s.addUserProcess(ctx, event, "")
	}

	switch actionType {
	case ActionTypeSetMeal:
		return s.setMeal(ctx, userID, event.ReplyToken, datas)
	case ActionTypeCheckDescript:
		return s.checkDescript(ctx, userID, event.ReplyToken, datas)
	case ActionTypeSetDescription:
		return s.setDescription(ctx, userID, event.ReplyToken)
	case ActionTypeCheckBasicInfo:
		if len(datas) > 0 && datas[0] == "y" {
			return s.checkBasicInfo(ctx, userID, event.ReplyToken)
		} else {
			s.removeRegisterProcess(userID)
			// Restart process
			return s.addUserProcess(ctx, event, "")
		}
	case ActionTypeDailyReport: // Not implemented in template yet, implementing JSON dump fallback
		return s.dailyReport(ctx, userID, user, event.ReplyToken)
	case ActionTypeJoinUs:
		return s.joinUs(ctx, event.ReplyToken)
	case ActionTypeGetMonth:
		return s.getMonth(ctx, userID, event.ReplyToken, datas)
	case ActionTypeSetReportStart:
		return s.setReportStart(ctx, userID, event.ReplyToken, datas)
	case ActionTypeSetReportEnd:
		return s.setReportEnd(ctx, userID, event.ReplyToken, datas)
	default:
		return s.replyText(ctx, event.ReplyToken, "unknown action")
	}
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
