package storage

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const dailyRecordTable = "daily_record"

func (p *Postgres) GetDailyRecord(ctx context.Context, q *query.Query) ([]models.DailyRecord, error) {
	builder := squirrel.Select("*").
		From(dailyRecordTable)
	builder = q.Where(builder)
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.DailyRecord])
}

func (p *Postgres) CreateDailyRecord(ctx context.Context, dailyRecord *models.DailyRecord) error {
	sql, args, err := squirrel.Insert(dailyRecordTable).
		Columns(
			"request_id",
			"line_id",
			"date",
			"total_water_ml",
			"total_sleep_hour",
			"breakfast_meals",
			"lunch_meals",
			"dinner_meals",
			"snack_meals").
		Values(dailyRecord.RequestID,
			dailyRecord.LineID,
			dailyRecord.Date,
			dailyRecord.TotalWaterMl,
			dailyRecord.TotalSleepHour,
			dailyRecord.BreakfastMeals,
			dailyRecord.LunchMeals,
			dailyRecord.DinnerMeals,
			dailyRecord.SnackMeals).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) UpdateDailyRecord(ctx context.Context, lineID string, date string, cols ...UpdateColumn) error {
	builder := squirrel.Update(dailyRecordTable).
		Where(squirrel.Eq{"line_id": lineID}).
		Where(squirrel.Eq{"date": date})
	for _, col := range cols {
		builder = builder.Set(string(col.ColumnName), col.Value)
	}
	sql, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}
