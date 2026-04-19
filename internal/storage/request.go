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
			"line_id",
			"request_type",
			"status",
			"data").
		Values(request.RequestID,
			request.LineID,
			request.RequestType,
			request.Status,
			request.Data).
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

func (p *Postgres) UpdateSendRequest(ctx context.Context, requestID string, vals ...UpdateColumn) error {
	builder := squirrel.Update(sendRequestsTable)
	for _, val := range vals {
		builder = builder.Set(string(val.ColumnName), val.Value)
	}
	sql, args, err := builder.
		Where(squirrel.Eq{"request_id": requestID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	_, err = p.sqlexer.Exec(ctx, sql, args...)
	return err
}
