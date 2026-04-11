package storage

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) UpsertUsage(ctx context.Context, usage *models.Usage) error {
	// if the line_id exists, update the usage
	// else insert the usage
	// this is a upsert operation on postgres
	sql, args, err := squirrel.Insert("used_tokens").
		Columns("line_id", "used_token").
		Values(usage.LineID, usage.Usage).
		Suffix("ON CONFLICT (line_id) DO UPDATE SET used_token = EXCLUDED.used_token").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) GetUsage(ctx context.Context, lineID string) (*models.Usage, error) {
	sql, args, err := squirrel.Select("*").
		From("used_tokens").
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
	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Usage])
}
