package bot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/cache"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
)

const (
	DefaultMaxDailyToken int64 = 120000
)

type Service struct {
	lineClient         *messaging_api.MessagingApiAPI
	blobClient         *messaging_api.MessagingApiBlobAPI
	store              *storage.Postgres
	cache              *cache.UserContext
	gpt                gpt.NutritionAPI
	logger             *zap.Logger
	maxDailyToken      int64
	AnalyzeMealFn      func(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, mealInfo *gpt.MealInfoWithImage, s *Service) error
	AnalyzeMealDailyFn func(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, dailyInfo *gpt.DailyInfo, s *Service) error
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
		gpt:        gpt,
		// can be override by option
		AnalyzeMealFn:      AnalyzeMeal,
		AnalyzeMealDailyFn: AnalyzeDailyMeal,
		maxDailyToken:      DefaultMaxDailyToken,
		logger:             zap.NewNop(),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	s.logger.Info("Service initialized", zap.Int64("maxDailyToken", s.maxDailyToken))
	return s, nil
}

func (s *Service) HandleEvent(ctx context.Context, event *linebot.Event) error {
	switch event.Type {
	case linebot.EventTypeMessage:
		switch message := event.Message.(type) {
		case *linebot.TextMessage:
			return s.handleTextMessage(ctx, event, message)
		case *linebot.ImageMessage:
			return s.handleImageMessage(ctx, event, message)
		}
	case linebot.EventTypePostback:
		return s.handlePostback(ctx, event)
	}
	return nil
}

func (s *Service) handleTextMessage(ctx context.Context, event *linebot.Event, message *linebot.TextMessage) error {
	userID := event.Source.UserID
	text := message.Text

	user, err := s.getUserInfo(ctx, userID)
	if err != nil {
		s.logger.Error("Error getting user info", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetUser.Error())
	}

	// Register logic
	if user == nil || s.cache.GetRegisterProcess(userID) {
		return s.addUserProcess(ctx, event, text)
	}

	// Registered user: Meal Logic
	meal := s.cache.GetMeal(userID)
	if meal == 0 {
		return s.replyText(ctx, event.ReplyToken, "請點擊下方選單開始上傳餐點")
	}

	s.cache.SetMealDescription(userID, text)

	mealName := getMealName(meal)

	vars := []template.Variable{
		{Name: "餐點", Value: mealName},
		{Name: "名稱", Value: text},
	}
	msg := template.GetCheckMsg("上傳餐點",
		vars,
		[]string{fmt.Sprintf("action=%s&data=y", ActionTypeCheckDescript.String()),
			fmt.Sprintf("action=%s&data=n", ActionTypeCheckDescript.String())})

	return s.replyFlex(ctx, event.ReplyToken, "upload img", msg)
}

func (s *Service) handleImageMessage(ctx context.Context, event *linebot.Event, message *linebot.ImageMessage) error {
	userID := event.Source.UserID

	user, err := s.getUserInfo(ctx, userID)
	if err != nil {
		s.logger.Error("Error getting user info", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetUser.Error())
	}
	// Register logic if user not found
	if user == nil {
		return s.addUserProcess(ctx, event, "")
	}

	meal := s.cache.GetMeal(userID)
	description := s.cache.GetMealDescription(userID)

	if meal == 0 || description == "" {
		msg := template.GetSetMealMsg()
		return s.replyFlex(ctx, event.ReplyToken, "set description", msg)
	}

	// Get Image Content
	content, err := s.blobClient.GetMessageContent(message.ID)
	if err != nil {
		s.logger.Error("Error getting image content", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrFaiedToGetImage.Error())
	}
	defer content.Body.Close()

	imgData, err := io.ReadAll(content.Body)
	if err != nil {
		s.logger.Error("Error reading image content", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}

	s.cache.SetPicture(userID, imgData)
	// Send request to AI
	uid := uuid.New()
	if err := s.store.CreateSendRequest(ctx, &models.SendRequest{
		RequestID:   uid.String(),
		RequestType: models.SendRequestTypeMeal,
		Status:      models.SendRequestStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}); err != nil {
		s.logger.Error("Error creating send request", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}

	mealInfo := &gpt.MealInfoWithImage{
		Image: imgData,
		MealInfo: gpt.MealInfo{
			Meal:        meal,
			Description: description,
			UserProfile: user.ToProfileString(),
		},
	}

	usedToken, err := s.store.GetUsage(ctx, userID)
	if err != nil {
		s.logger.Error("Error getting usage", zap.Error(err))
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrInternal.Error())
	}
	if time.Since(usedToken.UpdatedAt) > 24*time.Hour {
		usedToken.Usage = 0
	}
	if usedToken.Usage >= s.maxDailyToken {
		return s.replyText(ctx, event.ReplyToken, internalErrors.ErrOutOfDailyToken.Error())
	}

	go func() {
		if err := s.AnalyzeMealFn(ctx, uid, userID, usedToken.Usage, mealInfo, s); err != nil {
			s.logger.Error("AnalyzeMeal error", zap.Error(err))
			err = s.store.UpdateSendRequest(ctx, &models.SendRequest{
				RequestID: uid.String(),
				Status:    models.SendRequestStatusFailed,
				Error:     err.Error(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				s.logger.Error("UpdateSendRequest error", zap.Error(err))
			}
		}
	}()
	return s.replyText(ctx, event.ReplyToken, "AI 分析中，請稍後")
}

func (s *Service) addUserProcess(ctx context.Context, event *linebot.Event, text string) error {
	userID := event.Source.UserID

	if !s.cache.GetRegisterProcess(userID) {
		s.cache.SetRegisterProcess(userID, true)
		return s.replyText(ctx, event.ReplyToken, "歡迎使用營養師機器人，請依序輸入您的身高(公分)、體重(公斤)、年齡、性別(男/女)以完成註冊\n\n請輸入身高(公分) ex. 175.5")
	}

	if s.cache.GetHeight(userID) == 0 {
		h, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return s.replyText(ctx, event.ReplyToken, "身高格式錯誤，請重新輸入 ex. 175.5")
		}
		s.cache.SetHeight(userID, h)
		return s.replyText(ctx, event.ReplyToken, "請輸入體重(公斤) ex. 70.5")
	}

	if s.cache.GetWeight(userID) == 0 {
		w, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return s.replyText(ctx, event.ReplyToken, "體重格式錯誤，請重新輸入 ex. 70.5")
		}
		s.cache.SetWeight(userID, w)
		return s.replyText(ctx, event.ReplyToken, "請輸入年齡 ex. 25")
	}

	if s.cache.GetAge(userID) == 0 {
		a, err := strconv.Atoi(text)
		if err != nil {
			return s.replyText(ctx, event.ReplyToken, "年齡格式錯誤，請重新輸入 ex. 25")
		}
		s.cache.SetAge(userID, a)
		return s.replyText(ctx, event.ReplyToken, "請輸入性別(男/女)")
	}

	if s.cache.GetGender(userID) == 0 {
		t := strings.TrimSpace(text)
		var g models.Gender
		switch t {
		case "男":
			g = models.GenderMale
		case "女":
			g = models.GenderFemale
		default:
			return s.replyText(ctx, event.ReplyToken, "性別格式錯誤，請重新輸入(男/女)")
		}
		s.cache.SetGender(userID, int(g))

		// Confirm Msg
		genderStr := "男"
		if g == models.GenderFemale {
			genderStr = "女"
		}

		vars := []template.Variable{
			{Name: "性別", Value: genderStr},
			{Name: "身高", Value: fmt.Sprintf("%.1f 公分", s.cache.GetHeight(userID))},
			{Name: "體重", Value: fmt.Sprintf("%.1f 公斤", s.cache.GetWeight(userID))},
			{Name: "年齡", Value: fmt.Sprintf("%d 歲", s.cache.GetAge(userID))},
		}
		msg := template.GetCheckMsg("基本資料", vars, []string{"action=check_basic_info&data=y", "action=check_basic_info&data=n"})
		return s.replyFlex(ctx, event.ReplyToken, "basic info", msg)
	}

	return nil
}

func (s *Service) removeRegisterProcess(userID string) {
	s.cache.DeleteRegisterProcess(userID)
	s.cache.DeleteHeight(userID)
	s.cache.DeleteWeight(userID)
	s.cache.DeleteAge(userID)
	s.cache.DeleteGender(userID)
}

func (s *Service) getUserInfo(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.store.GetUserByLineID(ctx, userID)
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

func AnalyzeMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, mealInfo *gpt.MealInfoWithImage, s *Service) error {
	// AI Analysis
	start := time.Now()
	aiResp, usage, err := s.gpt.GetMealInfo(ctx, mealInfo)
	if usage > 0 {
		if err := s.store.UpsertUsage(ctx, &models.Usage{
			LineID:    userID,
			Usage:     usedToken + usage,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}); err != nil {
			s.logger.Error("DB Error", zap.Error(err))
		}
	}
	if err != nil {
		s.logger.Error("AI Error", zap.Error(err))
		msg := messaging_api.TextMessage{
			Text: internalErrors.ErrFailedToAnalyze.Error(),
		}

		if sendErr := s.sendMsg(ctx, userID, uid.String(), msg); sendErr != nil {
			log.Printf("Error sending message: %v", sendErr)
			err = errors.Join(err, sendErr)
		}
		return err
	}

	if !aiResp.IsFood {
		msg := messaging_api.TextMessage{
			Text: internalErrors.ErrNotFoodType.Error(),
		}
		sendErr := s.sendMsg(ctx, userID, uid.String(), msg)
		if sendErr != nil {
			s.logger.Error("Error sending message", zap.Error(sendErr))
			err = errors.Join(internalErrors.ErrNotFoodType, sendErr)
		}
		return err
	}

	history := getHistoryFromResp(userID, uid, mealInfo, aiResp)

	if err := s.store.CreateMealHistory(ctx, history); err != nil {
		s.logger.Error("DB Error", zap.Error(err))
		msg := messaging_api.TextMessage{
			Text: internalErrors.ErrInternal.Error(),
		}
		if sendErr := s.sendMsg(ctx, userID, uid.String(), msg); sendErr != nil {
			s.logger.Error("Error sending message", zap.Error(sendErr))
			err = errors.Join(err, sendErr)
		}
		return err
	}

	s.cache.DeleteMeal(userID)
	s.cache.DeleteMealDescription(userID)
	s.cache.DeletePicture(userID)

	cost := time.Since(start).Seconds()

	respMsg := template.GetAIMealResponseMsg(history, cost)
	msg := messaging_api.FlexMessage{
		AltText:  "AI Meal Response",
		Contents: respMsg,
	}
	if err := s.sendMsg(ctx, userID, uid.String(), msg); err != nil {
		s.logger.Error("Error sending message", zap.Error(err))
		return err
	}
	if err := s.store.UpdateSendRequest(ctx, &models.SendRequest{
		RequestID: uid.String(),
		Status:    models.SendRequestStatusSuccess,
		UpdatedAt: time.Now(),
	}); err != nil {
		s.logger.Error("UpdateSendRequest error", zap.Error(err))
	}

	return nil
}

func AnalyzeDailyMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, dailyInfo *gpt.DailyInfo, s *Service) error {
	// AI Analysis
	start := time.Now()
	aiResp, usage, err := s.gpt.GetMealDailyInfo(ctx, dailyInfo)
	if usage > 0 {
		if err := s.store.UpsertUsage(ctx, &models.Usage{
			LineID:    userID,
			Usage:     usedToken + usage,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}); err != nil {
			s.logger.Error("DB Error", zap.Error(err))
		}
	}
	if err != nil {
		s.logger.Error("AI Error", zap.Error(err))
		msg := messaging_api.TextMessage{
			Text: internalErrors.ErrFailedToAnalyze.Error(),
		}

		if sendErr := s.sendMsg(ctx, userID, uid.String(), msg); sendErr != nil {
			log.Printf("Error sending message: %v", sendErr)
			err = errors.Join(err, sendErr)
		}
		return err
	}

	cost := time.Since(start).Seconds()

	history := getDailyHistoryFromResp(userID, uid, dailyInfo, aiResp)

	if err := s.store.CreateMealDaily(ctx, history); err != nil {
		s.logger.Error("DB Error", zap.Error(err))
		msg := messaging_api.TextMessage{
			Text: internalErrors.ErrInternal.Error(),
		}
		if sendErr := s.sendMsg(ctx, userID, uid.String(), msg); sendErr != nil {
			s.logger.Error("Error sending message", zap.Error(sendErr))
			err = errors.Join(err, sendErr)
		}
		return err
	}

	respMsg := template.GetAIMealDailyMsg(history, cost)
	msg := messaging_api.FlexMessage{
		AltText:  "AI Meal Response",
		Contents: respMsg,
	}
	if err := s.sendMsg(ctx, userID, uid.String(), msg); err != nil {
		s.logger.Error("Error sending message", zap.Error(err))
		return err
	}
	if err := s.store.UpdateSendRequest(ctx, &models.SendRequest{
		RequestID: uid.String(),
		Status:    models.SendRequestStatusSuccess,
		UpdatedAt: time.Now(),
	}); err != nil {
		s.logger.Error("UpdateSendRequest error", zap.Error(err))
	}

	return nil
}

func getHistoryFromResp(userID string, uid uuid.UUID, mealInfo *gpt.MealInfoWithImage, resp *gpt.AIMealResponse) *models.MealHistory {
	// Calculate nutrition totals
	var (
		totalCals    float64
		totalProtein float64
		totalCarbs   float64
		totalFat     float64
		totalSodium  float64
	)
	for _, d := range resp.Dishes {
		totalCals += d.EstNutrition.CaloriesKcal
		totalProtein += d.EstNutrition.ProteinG
		totalCarbs += d.EstNutrition.CarbsG
		totalFat += d.EstNutrition.FatG
		totalSodium += d.EstNutrition.SodiumMg
	}

	history := &models.MealHistory{
		RequestID:     uid.String(),
		LineID:        userID,
		Meal:          models.Meal(mealInfo.Meal), // Assuming models.Meal is int-based enum
		Description:   mealInfo.Description,
		Photo:         mealInfo.Image,
		CaloriesKcal:  totalCals,
		ProteinG:      totalProtein,
		CarbsG:        totalCarbs,
		FatG:          totalFat,
		SodiumMg:      totalSodium,
		AIDescription: resp.ExplanationsZh,
		AIWarnings:    strings.Join(resp.Warnings, ";"),
		AISuggest:     resp.PersonalizedAdvice,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	return history
}

func getDailyHistoryFromResp(userID string, uid uuid.UUID, dailyInfo *gpt.DailyInfo, resp *gpt.AIDailyResponse) *models.MealDaily {
	meals := make([]models.Meal, 0)
	for _, meal := range dailyInfo.MealsToday {
		meals = append(meals, models.Meal(meal.Meal))
	}
	history := &models.MealDaily{
		RequestID:                    uid.String(),
		LineID:                       userID,
		Meals:                        models.GetMealsIntFromMeals(meals),
		TotalCaloriesKcal:            resp.DayTotals.CaloriesKcal,
		TotalProteinG:                resp.DayTotals.ProteinG,
		TotalCarbsG:                  resp.DayTotals.CarbsG,
		TotalFatG:                    resp.DayTotals.FatG,
		TotalSodiumMg:                resp.DayTotals.SodiumMg,
		ComplianceCalories:           resp.Compliance.Calories,
		ComplianceProtein:            resp.Compliance.Protein,
		ComplianceSodium:             resp.Compliance.Sodium,
		ComplianceDeltasCaloriesKcal: resp.Compliance.Deltas.CaloriesKcal,
		ComplianceDeltasProteinG:     resp.Compliance.Deltas.ProteinG,
		ComplianceDeltasSodiumMg:     resp.Compliance.Deltas.SodiumMg,
		Insights:                     resp.Insights,
		TodayCoaching:                resp.TodayCoaching,
		DataQualityIssues:            resp.DataQualityIssues,
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
	}
	return history
}
