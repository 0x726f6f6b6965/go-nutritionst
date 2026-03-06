package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxSqlExecutor interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	// pgxpool.Pool does not have "Prepare" method, but pgx.Conn and pgx.Tx do
	// However, pgx has automatic statement preparation and caching
	// https://github.com/jackc/pgx/issues/632
	// Prepare(ctx context.Context, name string, sql string) (*pgconn.StatementDescription, error)
}

type Postgres struct {
	// sqlexer is the stub to execute underlying db package functions
	// can be either pgxpool.Pool or pgx.Tx
	sqlexer  pgxSqlExecutor
	connPool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		sqlexer:  pool,
		connPool: pool,
	}
}
