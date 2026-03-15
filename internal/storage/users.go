package storage

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
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
			"age",
			"gender",
			"max_daily_token",
			"created_at",
			"updated_at").
		Values(user.LineID,
			user.Height,
			user.Weight,
			user.Age,
			user.Gender,
			user.MaxDailyToken,
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
