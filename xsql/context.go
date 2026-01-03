package xsql

import (
	"context"
	"database/sql"
)

type contextKey struct{}

func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, contextKey{}, tx)
}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(contextKey{}).(*sql.Tx)
	if tx == nil {
		return nil, false
	}

	return tx, ok
}

func ContextWithoutTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, nil)
}
