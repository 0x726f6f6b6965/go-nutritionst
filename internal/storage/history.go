package storage

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const (
	mealHistoryTable = "meal_history"
	mealDailyTable   = "meal_daily"
)

func (p *Postgres) CreateMealHistory(ctx context.Context, mealHistory *models.MealHistory, date string, logger *zap.Logger) error {
	// use begin transaction
	tx, err := p.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(context.Background()); err != nil && err != pgx.ErrTxClosed {
			logger.Error("failed to rollback transaction", zap.Error(err))
		}
	}()
	db := p.WithTx(tx)
	sql, args, err := squirrel.Insert(mealHistoryTable).
		Columns(
			"request_id",
			"line_id",
			"meal",
			"description",
			"photo",
			"calories_kcal",
			"protein_g",
			"carbs_g",
			"fat_g",
			"sodium_mg",
			"ai_description",
			"ai_warnings",
			"ai_suggest",
			"created_at",
			"updated_at").
		Values(mealHistory.RequestID,
			mealHistory.LineID,
			mealHistory.Meal,
			mealHistory.Description,
			mealHistory.Photo,
			mealHistory.CaloriesKcal,
			mealHistory.ProteinG,
			mealHistory.CarbsG,
			mealHistory.FatG,
			mealHistory.SodiumMg,
			mealHistory.AIDescription,
			mealHistory.AIWarnings,
			mealHistory.AISuggest,
			mealHistory.CreatedAt,
			mealHistory.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	// get meal daily
	sql, args, err = squirrel.Select("1").
		From(dailyRecordTable).
		Where(squirrel.Eq{"line_id": mealHistory.LineID}).
		Where(squirrel.Eq{"date": date}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	dailyRecords, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.DailyRecord])
	if err != nil {
		return err
	}
	if len(dailyRecords) == 0 {
		// create new daily record
		dailyRecord := models.DailyRecord{
			RequestID: uuid.New(),
			LineID:    mealHistory.LineID,
			Date:      date,
		}
		switch mealHistory.Meal {
		case models.MealBreakfast:
			dailyRecord.BreakfastMeals = 1
		case models.MealLunch:
			dailyRecord.LunchMeals = 1
		case models.MealDinner:
			dailyRecord.DinnerMeals = 1
		case models.MealSnack:
			dailyRecord.SnackMeals = 1
		}
		if err := db.CreateDailyRecord(ctx, &dailyRecord); err != nil {
			return err
		}
	} else {
		// update daily record
		dailyRecord := dailyRecords[0]
		cols := []UpdateColumn{}
		switch mealHistory.Meal {
		case models.MealBreakfast:
			dailyRecord.BreakfastMeals++
			cols = append(cols, UpdateColumn{
				ColumnName: DailyRecordBreakfastMeals,
				Value:      dailyRecord.BreakfastMeals,
			})
		case models.MealLunch:
			dailyRecord.LunchMeals++
			cols = append(cols, UpdateColumn{
				ColumnName: DailyRecordLunchMeals,
				Value:      dailyRecord.LunchMeals,
			})
		case models.MealDinner:
			dailyRecord.DinnerMeals++
			cols = append(cols, UpdateColumn{
				ColumnName: DailyRecordDinnerMeals,
				Value:      dailyRecord.DinnerMeals,
			})
		case models.MealSnack:
			dailyRecord.SnackMeals++
			cols = append(cols, UpdateColumn{
				ColumnName: DailyRecordSnackMeals,
				Value:      dailyRecord.SnackMeals,
			})
		}
		if err := db.UpdateDailyRecord(ctx, mealHistory.LineID, date, cols...); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (p *Postgres) GetMealHistory(ctx context.Context, query *query.Query) ([]models.MealHistory, error) {
	builder := squirrel.Select("*").
		From(mealHistoryTable)
	builder = query.Where(builder)
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.MealHistory])
}

func (p *Postgres) CreateMealDaily(ctx context.Context, mealDaily *models.MealDaily) error {
	sql, args, err := squirrel.Insert(mealDailyTable).
		Columns(
			"request_id",
			"line_id",
			"date",
			"breakfast_meals",
			"lunch_meals",
			"dinner_meals",
			"snack_meals",
			"total_calories_kcal",
			"total_protein_g",
			"total_carbs_g",
			"total_fat_g",
			"total_sodium_mg",
			"compliance_calories",
			"compliance_protein",
			"compliance_sodium",
			"compliance_deltas_calories_kcal",
			"compliance_deltas_protein_g",
			"compliance_deltas_sodium_mg",
			"insights",
			"today_coaching",
			"data_quality_issues",
			"created_at",
			"updated_at").
		Values(mealDaily.RequestID,
			mealDaily.LineID,
			mealDaily.Date,
			mealDaily.BreakfastMeals,
			mealDaily.LunchMeals,
			mealDaily.DinnerMeals,
			mealDaily.SnackMeals,
			mealDaily.TotalCaloriesKcal,
			mealDaily.TotalProteinG,
			mealDaily.TotalCarbsG,
			mealDaily.TotalFatG,
			mealDaily.TotalSodiumMg,
			mealDaily.ComplianceCalories,
			mealDaily.ComplianceProtein,
			mealDaily.ComplianceSodium,
			mealDaily.ComplianceDeltasCaloriesKcal,
			mealDaily.ComplianceDeltasProteinG,
			mealDaily.ComplianceDeltasSodiumMg,
			mealDaily.Insights,
			mealDaily.TodayCoaching,
			mealDaily.DataQualityIssues,
			mealDaily.CreatedAt,
			mealDaily.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) GetMealDaily(ctx context.Context, query *query.Query) ([]models.MealDaily, error) {
	builder := squirrel.Select("*").
		From(mealDailyTable)
	builder = query.Where(builder)
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.MealDaily])
}
