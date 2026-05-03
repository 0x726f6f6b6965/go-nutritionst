package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/0x726f6f6b6965/go-nutritionst/pkg/timezone"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	usersTable = "users"
	TimeDiff   = -5
)

func (p *Postgres) CreateUser(ctx context.Context, user *models.User) error {
	sql, args, err := squirrel.Insert(usersTable).
		Columns(
			"line_id",
			"height",
			"weight",
			"target_weight",
			"target_timeframe",
			"age",
			"gender",
			"max_daily_token",
			"morning_msg_sent",
			"evening_msg_sent",
			"breakfast_msg_sent",
			"lunch_msg_sent",
			"dinner_msg_sent",
			"created_at",
			"updated_at").
		Values(user.LineID,
			user.Height,
			user.Weight,
			user.TargetWeight,
			user.TargetTimeframe,
			user.Age,
			user.Gender,
			user.MaxDailyToken,
			user.MorningMsgSent,
			user.EveningMsgSent,
			user.BreakfastMsgSent,
			user.LunchMsgSent,
			user.DinnerMsgSent,
			user.CreatedAt,
			user.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) GetUserByLineID(ctx context.Context, lineID string) (*models.User, error) {
	sql, args, err := squirrel.Select("*").
		From(usersTable).
		Where(squirrel.Eq{"line_id": lineID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.User])
}

func (p *Postgres) UpdateUserTarget(ctx context.Context, lineID string, targetWeight float64, targetTimeframe int) error {
	sql, args, err := squirrel.Update(usersTable).
		Set("target_weight", targetWeight).
		Set("target_timeframe", targetTimeframe).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"line_id": lineID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) UpdateUserWeight(ctx context.Context, lineID string, weight float64) error {
	sql, args, err := squirrel.Update(usersTable).
		Set("weight", weight).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"line_id": lineID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) GetUsers(ctx context.Context, q *query.Query) ([]models.User, error) {
	builder := squirrel.Select("*").
		From(usersTable)
	builder = q.Where(builder)
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.User])
}

func (p *Postgres) UpdateUser(ctx context.Context, lineID string, vals ...UpdateColumn) error {
	builder := squirrel.Update(usersTable)
	for _, val := range vals {
		builder = builder.Set(string(val.ColumnName), val.Value)
	}
	sql, args, err := builder.Where(squirrel.Eq{"line_id": lineID}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) GetUsersWithNotEatMeal(ctx context.Context, q *query.Query, meal models.Meal) ([]models.User, error) {
	builder := squirrel.Select("*").
		From(usersTable)

	var sqlStr string
	switch meal {
	case models.MealBreakfast:
		sqlStr = fmt.Sprintf("SELECT line_id FROM %s WHERE breakfast_meals > 0 AND date = %s", dailyRecordTable, timezone.GetTaipeiDate())
	case models.MealLunch:
		sqlStr = fmt.Sprintf("SELECT line_id FROM %s WHERE lunch_meals > 0 AND date = %s", dailyRecordTable, timezone.GetTaipeiDate())
	case models.MealDinner:
		sqlStr = fmt.Sprintf("SELECT line_id FROM %s WHERE dinner_meals > 0 AND date = %s", dailyRecordTable, timezone.GetTaipeiDate())
	case models.MealSnack:
		sqlStr = fmt.Sprintf("SELECT line_id FROM %s WHERE snack_meals > 0 AND date = %s", dailyRecordTable, timezone.GetTaipeiDate())
	}

	q.AddFilter(squirrel.Expr(fmt.Sprintf("line_id NOT IN (%s)", sqlStr)))
	builder = q.Where(builder)
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[models.User])
}
