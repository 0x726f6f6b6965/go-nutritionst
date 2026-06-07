package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	internalErrors "github.com/0x726f6f6b6965/go-nutritionst/internal/errors"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/template"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/gpt"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/timezone"
	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"go.uber.org/zap"
)

func (s *Service) AnalyzeMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken *models.Usage, mealInfo *gpt.MealInfoWithImage) error {
	// AI Analysis
	start := time.Now()
	aiResp, usage, err := s.gpt.GetMealInfo(ctx, mealInfo)
	if err := s.updateUsage(ctx, userID, usedToken, usage); err != nil {
		s.logger.Error("Error updating usage", zap.Error(err))
		return err
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
	date := timezone.GetTaipeiDate()
	if err := s.store.CreateMealHistory(ctx, history, date, s.logger); err != nil {
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

func (s *Service) AnalyzeDailyMeal(ctx context.Context, uid uuid.UUID, userID string, usedToken *models.Usage, dailyInfo *gpt.DailyInfo) error {
	// AI Analysis
	start := time.Now()
	aiResp, usage, err := s.gpt.GetMealDailyInfo(ctx, dailyInfo)
	if err := s.updateUsage(ctx, userID, usedToken, usage); err != nil {
		s.logger.Error("Error updating usage", zap.Error(err))
		return err
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
	var (
		breakfastMeals int
		lunchMeals     int
		dinnerMeals    int
		snackMeals     int
	)
	for _, meal := range dailyInfo.MealsToday {
		switch meal.Meal {
		case int(models.MealBreakfast):
			breakfastMeals++
		case int(models.MealLunch):
			lunchMeals++
		case int(models.MealDinner):
			dinnerMeals++
		case int(models.MealSnack):
			snackMeals++
		}
	}
	history := &models.MealDaily{
		RequestID:                    uid.String(),
		LineID:                       userID,
		BreakfastMeals:               breakfastMeals,
		LunchMeals:                   lunchMeals,
		DinnerMeals:                  dinnerMeals,
		SnackMeals:                   snackMeals,
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

func (s *Service) AnalyzeBasicInfo(ctx context.Context, uid uuid.UUID, userID string, usedToken *models.Usage, basicInfo *gpt.BasicUserInfo) error {
	// AI Analysis
	aiResp, usage, err := s.gpt.GetTargetSuggestion(ctx, basicInfo)
	if err := s.updateUsage(ctx, userID, usedToken, usage); err != nil {
		s.logger.Error("Error updating usage", zap.Error(err))
		return err
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
		subUid := uuid.New()
		if err := s.store.CreateSendRequest(ctx, &models.SendRequest{
			RequestID:   subUid.String(),
			RequestType: models.SendRequestTypeBasicInfo,
			LineID:      userID,
			Status:      models.SendRequestStatusPending,
		}); err != nil {
			s.logger.Error("Error creating send request", zap.Error(err))
			return err
		}
		if err := s.sendStartMsg(ctx, userID, subUid.String()); err != nil {
			s.logger.Error("Error sending message", zap.Error(err))
			if sendErr := s.store.UpdateSendRequest(ctx, subUid.String(), storage.UpdateColumn{
				ColumnName: storage.SendRequestStatus,
				Value:      models.SendRequestStatusFailed,
			}, storage.UpdateColumn{
				ColumnName: storage.SendRequestFailReason,
				Value:      err.Error(),
			}); sendErr != nil {
				s.logger.Error("Error updating send request", zap.Error(sendErr))
			}
			return err
		}
		if err := s.store.UpdateSendRequest(ctx, subUid.String(), storage.UpdateColumn{
			ColumnName: storage.SendRequestStatus,
			Value:      models.SendRequestStatusSuccess,
		}); err != nil {
			s.logger.Error("UpdateSendRequest error", zap.Error(err))
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

func (s *Service) sendStartMsg(ctx context.Context, userID string, uid string) error {
	var startMsg messaging_api.TextMessage
	profile, err := s.lineClient.GetProfile(userID)
	if err != nil {
		s.logger.Error("Error getting profile", zap.Error(err))
		startMsg = messaging_api.TextMessage{
			Text: fmt.Sprintf("Hi,\n%s",
				template.DescriptionMsgStartUse.String(),
			),
		}
	} else {
		startMsg = messaging_api.TextMessage{
			Text: fmt.Sprintf("Hi %s,\n%s",
				profile.DisplayName,
				template.DescriptionMsgStartUse.String(),
			),
		}
	}
	if err := s.sendMsg(ctx, userID, uid, startMsg); err != nil {
		s.logger.Error("Error sending message", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) updateUsage(ctx context.Context, uid string, usedToken *models.Usage, usage int64) error {
	if usage > 0 {
		if err := s.store.UpsertUsage(ctx, &models.Usage{
			LineID:       uid,
			Usage:        usedToken.Usage + usage,
			LastUsedDate: timezone.GetTaipeiDate(),
		}); err != nil {
			s.logger.Error("DB Error", zap.Error(err))
		}
	}
	return nil
}
