package storage

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	mealHistoryTable = "meal_history"
	mealDailyTable   = "meal_daily"
)

func (p *Postgres) CreateMealHistory(ctx context.Context, mealHistory *models.MealHistory) error {
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
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
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
			"meals",
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
			mealDaily.Meals,
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
