package ai

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

func (s *Service) AnalyzeMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, mealInfo *gpt.MealInfoWithImage) error {
	// AI Analysis
	start := time.Now()
	aiResp, usage, err := s.gpt.GetMealInfo(ctx, mealInfo)
	if usage > 0 {
		if err := s.store.UpsertUsage(ctx, &models.Usage{
			LineID: userID,
			Usage:  usedToken + usage,
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
	if err := s.store.UpdateSendRequest(ctx, uid.String(), storage.UpdateColumn{
		ColumnName: storage.SendRequestStatus,
		Value:      models.SendRequestStatusSuccess,
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

func (s *Service) AnalyzeDailyMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, dailyInfo *gpt.DailyInfo) error {
	// AI Analysis
	start := time.Now()
	aiResp, usage, err := s.gpt.GetMealDailyInfo(ctx, dailyInfo)
	if usage > 0 {
		if err := s.store.UpsertUsage(ctx, &models.Usage{
			LineID: userID,
			Usage:  usedToken + usage,
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
	if err := s.store.UpdateSendRequest(ctx, uid.String(), storage.UpdateColumn{
		ColumnName: storage.SendRequestStatus,
		Value:      models.SendRequestStatusSuccess,
	}); err != nil {
		s.logger.Error("UpdateSendRequest error", zap.Error(err))
	}

	return nil
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

func (s *Service) AnalyzeBasicInfo(ctx context.Context, uid uuid.UUID, userID string, usedToken int64, basicInfo *gpt.BasicUserInfo) error {
	// AI Analysis
	aiResp, usage, err := s.gpt.GetTargetSuggestion(ctx, basicInfo)
	if usage > 0 {
		if err := s.store.UpsertUsage(ctx, &models.Usage{
			LineID: userID,
			Usage:  usedToken + usage,
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

	if aiResp.IsReasonable {
		msg := messaging_api.TextMessage{
			Text: aiResp.Suggestions,
		}
		if err := s.sendMsg(ctx, userID, uid.String(), msg); err != nil {
			s.logger.Error("Error sending message", zap.Error(err))
			return err
		}
	} else {
		respMsg := template.GetTargetMsg(aiResp, basicInfo.TargetWeight, basicInfo.TargetTimeframe)
		msg := messaging_api.FlexMessage{
			AltText:  "AI Suggestion",
			Contents: respMsg,
		}
		if err := s.sendMsg(ctx, userID, uid.String(), msg); err != nil {
			s.logger.Error("Error sending message", zap.Error(err))
			return err
		}
	}
	if err := s.store.UpdateSendRequest(ctx, uid.String(), storage.UpdateColumn{
		ColumnName: storage.SendRequestStatus,
		Value:      models.SendRequestStatusSuccess,
	}); err != nil {
		s.logger.Error("UpdateSendRequest error", zap.Error(err))
	}
	return nil
}
