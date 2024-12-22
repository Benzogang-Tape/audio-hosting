package postgres

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pg"
)

func (p *PgStorage) Begin(ctx context.Context) (PgTx, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return PgTx{}, err //nolint:wrapcheck
	}

	return PgTx{
		tx:      tx,
		Queries: New(tx),
	}, nil
}

// You can't commit a database, so it does nothing.
func (*PgStorage) Commit(context.Context) error {
	return nil
}

// You can't rollback a database, so it does nothing.
func (*PgStorage) Rollback(context.Context) error {
	return nil
}

type PgTx struct {
	tx pg.Tx
	*Queries
}

func (t PgTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx) //nolint:wrapcheck
}

func (t PgTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx) //nolint:wrapcheck
}

// You can't use Begin twice, so this is a workaround.
func (t PgTx) Begin(context.Context) (PgTx, error) {
	return t, nil
}
