package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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

// PgxIface is an interface for pgx abstraction
type PgxIface interface {
	pgxSqlExecutor
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
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

func (p *Postgres) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return p.sqlexer.Exec(ctx, sql, args...)
}

func (p *Postgres) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.sqlexer.Query(ctx, sql, args...)
}

func (p *Postgres) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.sqlexer.QueryRow(ctx, sql, args...)
}

func (p *Postgres) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.sqlexer.SendBatch(ctx, b)
}

func (p *Postgres) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.connPool.BeginTx(ctx, txOptions)
}

func (p *Postgres) NewBatch() *pgx.Batch {
	return &pgx.Batch{}
}

// WithTx creates a new Postgres with its sqlexer bound on a TX handle
func (p *Postgres) WithTx(tx pgx.Tx) *Postgres {
	return &Postgres{
		sqlexer:  tx,
		connPool: p.connPool,
	}
}

// ExecTx wraps the sql transaction procedure of pgx package.
// targetTxFn defines the DB operations to be executed in atomic fasion
func (p *Postgres) ExecTx(ctx context.Context, txOptions pgx.TxOptions, targetTxFn func(*Postgres) error) error {
	if p.sqlexer != p.connPool {
		// already in transaction, pass control to it
		return targetTxFn(p)
	}

	tx, err := p.connPool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback(context.Background())
		if rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			zap.S().Error(
				"tx rollback error",
				zap.NamedError("execErr", err),
				zap.NamedError("rollbackErr", rollbackErr),
			)
		}
	}()

	err = targetTxFn(p.WithTx(tx))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) Close() {
	p.connPool.Close()
}
