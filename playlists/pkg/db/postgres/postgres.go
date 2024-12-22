package postgres

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // it is needed for migrations to work

	"dev.gaijin.team/go/golib/e"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	UserName   string `env:"POSTGRES_USER" env-default:"playlists_user" yaml:"username"`
	Password   string `env:"POSTGRES_PASSWORD" env-default:"hard_password1234" yaml:"password"`
	Host       string `env:"POSTGRES_HOST" env-default:"localhost" yaml:"host"`
	Port       int    `env:"POSTGRES_PORT" env-default:"5432" yaml:"port"`
	DBName     string `env:"POSTGRES_DB" env-default:"DB" yaml:"dbName"`
	SSLEnabled string `env:"POSTGRES_SSL_ENABLED" env-default:"disable" yaml:"sslEnabled"`
}

type DB struct {
	Pool *pgxpool.Pool
}

func New(config Config, migrations fs.FS) (*DB, error) {
	conn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", //nolint:nosprintfhostport
		config.UserName, config.Password, config.Host, config.Port, config.DBName, config.SSLEnabled)

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, e.NewFrom("creating source", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, conn)
	if err != nil {
		return nil, e.NewFrom("migrate new", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, e.NewFrom("applying migrations", err)
	}

	p, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, e.NewFrom("creating pool", err)
	}

	return &DB{Pool: p}, nil
}

func (d DB) Begin(ctx context.Context) (Tx, error) {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return Tx{}, err //nolint:wrapcheck
	}

	return Tx{tx: tx}, nil
}

func (d DB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (Tx, error) {
	tx, err := d.Pool.BeginTx(ctx, txOptions)
	if err != nil {
		return Tx{}, err //nolint:wrapcheck
	}

	return Tx{tx: tx}, nil
}

func (d DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tag, err := d.Pool.Exec(ctx, query, args...)
	return tag, wrapError(err)
}

func (d DB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	rows, err := d.Pool.Query(ctx, query, args...)
	return rows, wrapError(err)
}

func (d DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	row := d.Pool.QueryRow(ctx, query, args...)
	return rowWrapped{rw: row}
}

type rowWrapped struct {
	rw pgx.Row
}

func (r rowWrapped) Scan(dest ...any) error {
	err := r.rw.Scan(dest...)
	return wrapError(err)
}
