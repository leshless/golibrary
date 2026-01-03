package xsql

import (
	"context"
	"database/sql"
)

type Client interface {
	Ping(ctx context.Context) error

	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row

	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)

	Close() error
}

type client struct {
	db *sql.DB
}

var _ Client = (*client)(nil)

func NewClient(db *sql.DB) *client {
	return &client{
		db: db,
	}
}

func (c *client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *client) Close() error {
	return c.db.Close()
}

func (c *client) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.ExecContext(ctx, query, args...)
	}

	return c.db.ExecContext(ctx, query, args...)
}

func (c *client) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.PrepareContext(ctx, query)
	}

	return c.db.PrepareContext(ctx, query)
}

func (c *client) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.QueryContext(ctx, query, args...)
	}

	return c.db.QueryContext(ctx, query, args...)
}

func (c *client) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx, ok := TxFromContext(ctx); ok {
		return tx.QueryRowContext(ctx, query, args...)
	}

	return c.db.QueryRowContext(ctx, query, args...)
}

func (c *client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if tx, ok := TxFromContext(ctx); ok {
		return tx, nil
	}

	return c.db.BeginTx(ctx, opts)
}
