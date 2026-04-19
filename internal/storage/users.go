package storage

import (
	"context"
	"time"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/query"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	usersTable = "users"
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
