package pg

import (
	"context"
	"errors"
	"io/fs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // it is needed for migrations to work

	"dev.gaijin.team/go/golib/e"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db *pgxpool.Pool
}

func Connect(ctx context.Context, conn string, migrations fs.FS) (Database, error) {
	db, err := pgxpool.New(ctx, conn)
	if err != nil {
		return Database{}, e.NewFrom("connecting to pg database", err)
	}

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return Database{}, e.NewFrom("creating source", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, conn)
	if err != nil {
		return Database{}, e.NewFrom("migrate new", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return Database{}, e.NewFrom("applying migrations", err)
	}

	return Database{
		db: db,
	}, nil
}

func (d Database) Close() error {
	d.db.Close()
	return nil
}

func (d Database) Begin(ctx context.Context) (Tx, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return Tx{}, err //nolint:wrapcheck
	}

	return Tx{tx: tx}, nil
}

func (d Database) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, txOptions)
	if err != nil {
		return Tx{}, err //nolint:wrapcheck
	}

	return Tx{tx: tx}, nil
}

func (d Database) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tag, err := d.db.Exec(ctx, query, args...)
	return tag, wrapError(err)
}

func (d Database) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	rows, err := d.db.Query(ctx, query, args...)
	return rows, wrapError(err)
}

func (d Database) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	row := d.db.QueryRow(ctx, query, args...)
	return rowWrapped{rw: row}
}

type rowWrapped struct {
	rw pgx.Row
}

func (r rowWrapped) Scan(dest ...any) error {
	err := r.rw.Scan(dest...)
	return wrapError(err)
}
