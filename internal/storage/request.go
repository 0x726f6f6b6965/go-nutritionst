package storage

import (
	"context"

	"github.com/0x726f6f6b6965/go-nutritionst/internal/storage/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

const (
	sendRequestsTable = "send_requests"
)

func (p *Postgres) CreateSendRequest(ctx context.Context, request *models.SendRequest) error {
	sql, args, err := squirrel.Insert(sendRequestsTable).
		Columns(
			"request_id",
			"request_type",
			"status",
			"data",
			"created_at",
			"updated_at").
		Values(request.RequestID,
			request.RequestType,
			request.Status,
			request.Data,
			request.CreatedAt,
			request.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}

func (p *Postgres) GetSendRequestByRequestID(ctx context.Context, requestID string) (*models.SendRequest, error) {
	sql, args, err := squirrel.Select("*").
		From(sendRequestsTable).
		Where(squirrel.Eq{"request_id": requestID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := p.sqlexer.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.SendRequest])
}

func (p *Postgres) UpdateSendRequest(ctx context.Context, request *models.SendRequest) error {
	sql, args, err := squirrel.Update(sendRequestsTable).
		Set("status", request.Status).
		Set("error", request.Error).
		Set("updated_at", request.UpdatedAt).
		Where(squirrel.Eq{"request_id": request.RequestID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}
