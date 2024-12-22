package postgres

import (
	"context"
	"dev.gaijin.team/go/golib/e"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Tx struct {
	tx pgx.Tx
}

func (d Tx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tag, err := d.tx.Exec(ctx, query, args...)
	return tag, wrapError(err)
}

func (d Tx) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	rows, err := d.tx.Query(ctx, query, args...)
	return rows, wrapError(err)
}

func (d Tx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	row := d.tx.QueryRow(ctx, query, args...)
	return rowWrapped{rw: row}
}

func (d Tx) Commit(ctx context.Context) error {
	err := d.tx.Commit(ctx)
	if err != nil {
		return e.NewFrom("commit tx", err)
	}

	return nil
}

func (d Tx) Rollback(ctx context.Context) error {
	err := d.tx.Rollback(ctx)
	if err != nil {
		return e.NewFrom("rollback tx", err)
	}

	return nil
}
